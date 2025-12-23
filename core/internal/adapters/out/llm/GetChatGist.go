package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const maxGistLength = 3900

// GetChatGist выполняет запрос к LLM - сценарий getChatGistFlow.
func (s *GenkitService) GetChatGist(ctx context.Context, messages []model.Message) ([]string, error) {

	log := slog.With("func", "llm.GetChatGist")
	log.Debug("get chat gist start", slog.Int("message count", len(messages)))

	ctxFlow, cancel := context.WithTimeout(ctx, s.flowTimeout)
	defer cancel()

	gist, err := s.getChatGistFlow.Run(ctxFlow, &chat{messages}) // Выполняем сценарий
	if err != nil {
		return nil, err
	}

	for i := range gist {
		log.Debug("get chat gist success", slog.Int("batch number", i), slog.String("chat gist", gist[i]))
	}

	return gist, nil
}

// defineGetChatGistFlow определяет сценарий для генерации краткого пересказа чата.
//
//nolint:gocognit
func (s *GenkitService) defineGetChatGistFlow() {

	// В промпте задается ограничение в maxGistLength (3900) символов. Так как Telegram ограничивает сообщение 4096 символами.
	prompt := `Тебе дана история сообщений из чата в Telegram. 
Каждое сообщение содержит текст, время отправки, ID отправителя и служебную информацию (например, было ли оно отредактировано, переслано или является ответом). 
Составь краткий и связный пересказ этой беседы. Сосредоточься на главных событиях, важных решениях, вопросах и ответах. 
Технические детали (например, метки «переслано» или «отредактировано») учитывай только если они влияют на смысл. 
Пересказ должен быть нейтральным, легко читаемым и помогать человеку быстро понять суть обсуждения. 
Пересказ должен быть меньше 3900 символов. 
История сообщений (в хронологическом порядке): {{messages}} 
Приведи только пересказ, без пояснений и дополнительного форматирования.`

	// Определяем простой запрос(prompt) getChatGistPrompt
	getChatGistPrompt := genkit.DefinePrompt(s.g, "getChatGistPrompt",
		ai.WithPrompt(prompt),                    // запрос
		ai.WithInputType(chat{}),                 // входные данные
		ai.WithOutputFormat(ai.OutputFormatText), // выходные данные
	)

	// Определяем сценарий(flow) getChatGistFlow
	s.getChatGistFlow = genkit.DefineFlow(s.g, "getChatGistFlow", func(ctx context.Context, input *chat) ([]string, error) {

		log := slog.With("func", "getChatGistFlow")
		// Разбивка сообщений на батчи, размером = contextWindow - driftPercent токенов.
		from := 0                 // начало батча
		to := 0                   // конец батча
		gist := make([]string, 0) // результат

		for {
			batchSize := 0
			for batchSize < (s.contextWindow-(s.contextWindow*s.driftPercent/100))*s.symbolPerToken && to < len(input.Messages) { // Ищем конец батча, укладывающегося в контекстное окно
				jsonData, errJ := json.Marshal(input.Messages[to])
				if errJ != nil {
					return nil, fmt.Errorf("marshal json message error: %w", errJ)
				}
				batchSize += len(jsonData)
				to++
			}

			log.Debug("Get chat gist batch",
				slog.Int("batch size (symbols)", batchSize),
				slog.Int("context window", s.contextWindow),
				slog.Int("batch from", from),
				slog.Int("batch to", to),
				slog.Int("all messages count", len(input.Messages)))

			batch := chat{
				Messages: input.Messages[from:to],
			}

			// выполняем простой запрос с Retry wrapper для обработки 429
			resp, err := retryPrompt(ctx, getChatGistPrompt, batch, log)
			if err != nil {
				return nil, fmt.Errorf("getChatGistFlow.getChatGistPrompt: %w", err)
			}

			// Ограничиваем длину сообщения.
			if len(resp.Text()) > maxGistLength {
				log.Warn("gist is too long, crop it", slog.Int("length", len(resp.Text())))
				gist = append(gist, resp.Text()[:maxGistLength]+"\ncropped!\n") // сохраняем суть сообщений текущего батча, обрезаем по длине, если превышает максимальную длину телеграм сообщения.
			} else {
				gist = append(gist, resp.Text()) // сохраняем суть сообщений текущего батча
			}

			if to == len(input.Messages) { // Прерываем цикл, после обработки всех сообщений
				break
			}

			from = to // Сдвигаем курсор батча
		}

		return gist, nil
	})
}

package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const maxGistLength = 3900

// GetChatGist выполняет запрос к LLM - сценарий getChatGistFlow.
func (s *GenkitService) GetChatGist(ctx context.Context, messages []model.Message) ([]model.BatchGist, error) {

	log := slog.With("func", "llm.GetChatGist")
	log.Debug("get chat gist start", slog.Int("message count", len(messages)))

	defer func() {
		if r := recover(); r != nil {
			log.Error("get chat gist panic", slog.Any("panic", r))
		}
	}()

	ctxFlow, cancel := context.WithTimeout(ctx, s.flowTimeout)
	defer cancel()

	gist, err := s.getChatGistFlow.Run(ctxFlow, &chat{messages}) // Выполняем сценарий
	if err != nil {
		return nil, err
	}

	for i := range gist {
		log.Debug("get chat gist success",
			slog.Int("batch number", i),
			slog.String("chat gist", gist[i].Gist),
			slog.Int("last message id", gist[i].LastMessageID))
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
	Составь КРАТКИЙ и связный пересказ чата (МАКСИМУМ 3900 символов, цель — 3000).
	Сосредоточься на главных событиях, важных решениях, вопросах и ответах.
	Технические детали (например, метки «переслано» или «отредактировано») учитывай только если они влияют на смысл.
	Пересказ должен быть нейтральным, легко читаемым и помогать человеку быстро понять суть обсуждения.
	История сообщений (в хронологическом порядке): {{messages}}
	Приведи только пересказ, без пояснений и дополнительного форматирования.`

	// Определяем простой запрос(prompt) getChatGistPrompt
	getChatGistPrompt := genkit.DefinePrompt(s.g, "getChatGistPrompt",
		ai.WithPrompt(prompt),                    // запрос
		ai.WithInputType(chat{}),                 // входные данные
		ai.WithOutputFormat(ai.OutputFormatText), // выходные данные
		ai.WithConfig(s.config),                  // В конфигурации также задается ограничение в output токенах, и прочие параметры работы llm
	)

	// Определяем сценарий(flow) getChatGistFlow
	s.getChatGistFlow = genkit.DefineFlow(s.g, "getChatGistFlow", func(ctx context.Context, input *chat) ([]model.BatchGist, error) {

		log := slog.With("func", "getChatGistFlow")
		// Разбивка сообщений на батчи, размером = contextWindow - driftPercent токенов.
		from := 0                          // начало батча
		to := 0                            // конец батча
		gist := make([]model.BatchGist, 0) // результат

		for {
			batchSize := 0
			for batchSize < (s.contextWindow-(s.contextWindow*s.driftPercent/100))*s.symbolPerToken && // Ищем конец батча, укладывающегося в контекстное окно
				to < len(input.Messages) && // Ограничение по количеству сообщений
				(to-from < s.messagesPerBatch) { // Ограничение по количеству сообщений в батче

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

			log.Debug("ответ от llm", slog.Any("resp.Text()", resp.Text()))

			lastMessageID := to // ID последнего сообщения в батче
			if lastMessageID == len(input.Messages) {
				lastMessageID--
			}

			// Ограничиваем длину сообщения. TODO вынести в логику телеграм клиента?
			if len(resp.Text()) > maxGistLength {
				log.Warn("gist is too long, crop it", slog.Int("length", len(resp.Text()))) // TODO исправить логику на работу с unicode
				gist = append(gist, model.BatchGist{
					Gist:          resp.Text()[:maxGistLength] + "\n\ncropped " + strconv.Itoa(len(resp.Text())-maxGistLength) + "!",
					LastMessageID: input.Messages[lastMessageID].ID,
					MessageCount:  to - from,
				}) // сохраняем суть сообщений текущего батча, обрезаем по длине, если превышает максимальную длину телеграм сообщения.
			} else {
				gist = append(gist, model.BatchGist{
					Gist:          resp.Text(),
					LastMessageID: input.Messages[lastMessageID].ID,
					MessageCount:  to - from,
				}) // сохраняем суть сообщений текущего батча
			}

			if to == len(input.Messages) { // Прерываем цикл, после обработки всех сообщений
				break
			}

			log.Debug("move to next batch", slog.Int("from", from), slog.Int("to", to), slog.Int("messages in batch", to-from), slog.Int("message count", len(input.Messages)))

			from = to // Сдвигаем курсор батча

		}

		log.Debug("getChatGistFlow success", slog.Int("gist count", len(gist)))

		return gist, nil
	})
}

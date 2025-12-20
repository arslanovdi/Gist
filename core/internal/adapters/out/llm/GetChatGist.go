package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const driftPercent = 20  // процент отклонения от заданного контекстного окна (в минус), так как количество токенов можно посчитать только приблизительно.
const symbolPerToken = 2 // 1 токен ~ 2 русских символа. Расчет приблизительный, так как неизвестно как работают токенизаторы различных LLM.

// GetChatGist выполняет запрос к LLM - сценарий getChatGistFlow.
func (s *GenkitService) GetChatGist(ctx context.Context, messages []model.Message) (string, error) {

	log := slog.With("func", "llm.GetChatGist")
	log.Debug("get chat gist start", slog.Int("message count", len(messages)))

	ctxFlow, cancel := context.WithTimeout(ctx, s.flowTimeout)
	defer cancel()

	// Разбивка сообщений на батчи, размером = contextWindow - driftPercent токенов.
	// TODO вынести логику разбивки на батчи во getChatGistFlow?
	from := 0                 // начало батча
	to := 0                   // конеч батча
	gist := strings.Builder{} // результат

	for {
		batchSize := 0
		for batchSize < (s.contextWindow-(s.contextWindow*driftPercent/100))*symbolPerToken && to < len(messages) { // Ищем конец батча, укладывающегося в контекстное окно
			jsonData, errJ := json.Marshal(messages[to])
			if errJ != nil {
				return "", fmt.Errorf("marshal json message error: %w", errJ)
			}
			batchSize += len(jsonData)
			to++
		}

		log.Debug("Get chat gist batch",
			slog.Int("batch size (symbols)", batchSize),
			slog.Int("context window", s.contextWindow),
			slog.Int("batch from", from),
			slog.Int("batch to", to),
			slog.Int("all messages count", len(messages)))

		resp, err := s.getChatGistFlow.Run(ctxFlow, &chat{messages[from:to]}) // Отправляем батч сообщений
		if err != nil {
			return "", err
		}

		gist.WriteString(resp) // сохраняем суть сообщений текущего батча

		if to == len(messages) { // Прерываем цикл, после обработки всех сообщений
			break
		}

		from = to // Сдвигаем курсор батча

	}

	log.Debug("get chat gist success", slog.String("chat gist", gist.String()))

	return gist.String(), nil
}

// defineGetChatGistFlow определяет сценарий для генерации краткого пересказа чата.
func (s *GenkitService) defineGetChatGistFlow() {

	// В промте задается ограничение в 4000 символов. Так как Telegram ограничивает сообщение 4096 символами.
	// TODO подумать над реализацией, что делать если много батчей сообщений. Результат будет очень большим.
	// анализировать батч, выводить результат пользователю и ждать нажатия кнопки далее, в промежутке сохранять номер необработанного сообщения в слайсе. И по нажатию на кнопку пометить прочитанным отмечать сообщения до этого номера (только обработанные сообщения)
	prompt := `Тебе дана история сообщений из чата в Telegram. 
Каждое сообщение содержит текст, время отправки, ID отправителя и служебную информацию (например, было ли оно отредактировано, переслано или является ответом). 
Составь краткий и связный пересказ этой беседы. Сосредоточься на главных событиях, важных решениях, вопросах и ответах. 
Технические детали (например, метки «переслано» или «отредактировано») учитывай только если они влияют на смысл. 
Пересказ должен быть нейтральным, легко читаемым и помогать человеку быстро понять суть обсуждения. 
Пересказ должен быть не более 4000 символов. 
История сообщений (в хронологическом порядке): {{messages}} 
Приведи только пересказ, без пояснений и дополнительного форматирования.`

	// Старайся уложиться в 150 слов, если беседа не слишком объёмная.
	// Определяем простой запрос(prompt) getChatGistPrompt
	getChatGistPrompt := genkit.DefinePrompt(s.g, "getChatGistPrompt",
		ai.WithPrompt(prompt),                    // запрос
		ai.WithInputType(chat{}),                 // входные данные
		ai.WithOutputFormat(ai.OutputFormatText), // выходные данные	TODO возвращать текст + аудио
	)

	// Определяем сценарий(flow) getChatGistFlow
	s.getChatGistFlow = genkit.DefineFlow(s.g, "getChatGistFlow", func(ctx context.Context, input *chat) (string, error) {
		// выполняем простой запрос
		resp, err := getChatGistPrompt.Execute(ctx,
			ai.WithInput(input), // входные данные
		)

		// TODO генерация аудио

		if err != nil {
			return "", err
		}
		return resp.Text(), nil
	})
}

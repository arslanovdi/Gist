package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/openai/openai-go"
)

// Тип входных данных для запроса к LLM.
type chat struct {
	Messages []model.Message `json:"messages"`
}

// GenerateChatGist выполняет запрос к LLM - сценарий generateChatGistStreamingFlow. callback - функция для оповещения пользователя о процессе выполнения.
func (s *GenkitService) GenerateChatGist(ctx context.Context, messages []model.Message, callback func(message string, progress int, llm bool)) ([]model.BatchGist, error) {

	log := slog.With("func", "llm.GenerateChatGist")
	log.Debug("get chat gist start", slog.Int("message count", len(messages)))

	defer func() {
		if r := recover(); r != nil {
			log.Error("get chat gist panic", slog.Any("panic", r))
		}
	}()

	ctxFlow, cancel := context.WithTimeout(ctx, s.flowTimeout)
	defer cancel()

	streamIter := s.generateChatGistStreamingFlow.Stream(ctxFlow, &chat{messages}) // Обработка Streaming Flow. С пошаговым оповещением пользователя о ходе процесса.
	var gist []model.BatchGist
	streamIter(func(value *core.StreamingFlowValue[[]model.BatchGist, *int], err error) bool {
		if err != nil {
			log.Error("stream error", slog.Any("error", err))
			return false
		}
		if value.Stream != nil {
			callback("⏳ Генерируем пересказ...", *value.Stream, true)
			log.Debug("flow step", slog.Int("progress", *value.Stream)) // уведомления пользователю value.Stream - % завершения
		}
		if value.Done {
			gist = value.Output // окончательный ответ streaming flow
		}
		return !value.Done // продолжать пока не Done
	})

	for i := range gist {
		log.Debug("get chat gist success",
			slog.Int("batch number", i),
			slog.String("chat gist", gist[i].Gist),
			slog.Int("last message id", gist[i].LastMessageID))
	}

	return gist, nil
}

// defineGenerateChatGistFlow определяет сценарий для генерации краткого пересказа чата.
//
//nolint:gocognit
func (s *GenkitService) defineGenerateChatGistFlow() {

	config := &openai.ChatCompletionNewParams{ //конфигурация для OpenRouter provider (OpenAI compatible), для других провайдеров нужно изменять!
		// Основные параметры
		Temperature:         openai.Float(0.1), // (0.0 - 2.0) Стабильность (низкие значения) / Креативность (высокие значения)
		MaxCompletionTokens: openai.Int(2000),  // Максимальное количество токенов в генерируемом ответе (output) Если больше, то ответ будет обрезан.
		TopP:                openai.Float(0.9), // Ограничивает выбор токенов топ-N% вероятностей. 0.9 — хороший баланс компактности и естественности.
		// Штрафы за повторения
		FrequencyPenalty: openai.Float(0.8), // Штрафует часто повторяющиеся слова (-2.0 до 2.0). Положительные значения → разнообразие текста.
		PresencePenalty:  openai.Float(0.6), // Штрафует любые повторяющиеся темы/сущности. 0.2-0.6 → фокус на новых идеях.
		// Поведение и формат
		/*N: openai.Int(1), // Количество вариантов ответа (по умолчанию 1)
		Stop: openai.ChatCompletionNewParamsStopUnion{ // Стоп-сигналы. Например: Stop: openai.F("Пересказ:") — остановка после ключевого слова.
			OfString:      param.Opt[string]{},
			OfStringArray: nil,
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{ // принуждает JSON-ответ.
			OfText:       nil,
			OfJSONSchema: nil,
			OfJSONObject: nil,
		},*/
	}

	// В промпте задается ограничение в maxGistLength (3900) символов. Так как Telegram ограничивает сообщение 4096 символами.
	prompt := `Тебе дана история сообщений из чата в Telegram.
	Каждое сообщение содержит текст, время отправки, ID отправителя и служебную информацию (например, было ли оно отредактировано, переслано или является ответом).
	Составь КРАТКИЙ и связный пересказ чата (длинной 3000-4000 символов, цель — 3000).
	Сосредоточься на главных событиях, важных решениях, вопросах и ответах.
	Технические детали (например, метки «переслано» или «отредактировано») учитывай только если они влияют на смысл.
	Пересказ должен быть нейтральным, легко читаемым и помогать человеку быстро понять суть обсуждения.
	История сообщений (в хронологическом порядке): {{messages}}
	Приведи только пересказ, без пояснений и дополнительного форматирования.`

	// Определяем простой запрос(prompt) generateChatGistPrompt
	generateChatGistPrompt := genkit.DefinePrompt(s.g, "generateChatGistPrompt",
		ai.WithPrompt(prompt),                    // запрос
		ai.WithInputType(chat{}),                 // входные данные
		ai.WithOutputFormat(ai.OutputFormatText), // выходные данные
		ai.WithConfig(config),                    // В конфигурации также задается ограничение в output токенах, и прочие параметры работы llm
		ai.WithModelName(s.DefaultTextModel),     // Используем дефолтную модель провайдер по умолчанию, заданного в конфигурации
	)

	// Определяем потоковый	 сценарий(streaming flow) generateChatGistStreamingFlow
	s.generateChatGistStreamingFlow = genkit.DefineStreamingFlow(s.g, "generateChatGistStreamingFlow",
		func(ctx context.Context, input *chat, cb func(ctx context.Context, percent *int) error) ([]model.BatchGist, error) {

			log := slog.With("func", "generateChatGistStreamingFlow")
			// Разбивка сообщений на батчи, размером = contextWindow - driftPercent токенов.
			from := 0                          // начало батча
			to := 0                            // конец батча
			gist := make([]model.BatchGist, 0) // результат
			messageProcessed := 0              // Счетчик обработанных сообщений
			progress := 0

			_ = cb(ctx, &progress) // show processing to user

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
				resp, err := retryPrompt(ctx, generateChatGistPrompt, batch, log)
				if err != nil {
					return nil, fmt.Errorf("getChatGistFlow.getChatGistPrompt: %w", err)
				}

				messageProcessed += to - from
				progress = messageProcessed * 100 / len(input.Messages)
				_ = cb(ctx, &progress) // show processing to user

				log.Debug("ответ от llm", slog.Any("resp.Text()", resp.Text()))

				lastMessageID := to // ID последнего сообщения в батче
				if lastMessageID == len(input.Messages) {
					lastMessageID--
				}

				gist = append(gist, model.BatchGist{
					Gist:            resp.Text(),
					LastMessageID:   input.Messages[lastMessageID].ID,
					LastMessageData: input.Messages[lastMessageID].Timestamp,
					MessageCount:    to - from, // кол-во обработанных сообщений в батче, учитывая и пропущенные (пустые, системные и т.п.)
				}) // сохраняем суть сообщений текущего батча

				if to == len(input.Messages) { // Прерываем цикл, после обработки всех сообщений
					break
				}

				log.Debug("move to next batch",
					slog.Int("from", from),
					slog.Int("to", to),
					slog.Int("messages in batch", to-from),
					slog.Int("message count", len(input.Messages)))

				from = to // Сдвигаем курсор батча

			}

			log.Debug("getChatGistFlow success", slog.Int("gist count", len(gist)))

			return gist, nil
		})
}

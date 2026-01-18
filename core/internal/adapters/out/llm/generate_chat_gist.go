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
	var errI error
	streamIter(func(value *core.StreamingFlowValue[[]model.BatchGist, *int], err error) bool {
		if err != nil {
			log.Error("stream error", slog.Any("error", err))
			errI = err
			return false
		}
		if value.Stream != nil {
			callback("⏳ Генерируем пересказ...", *value.Stream, true)
			log.Debug("flow step", slog.Int("progress", *value.Stream)) // уведомления пользователю value.Stream - % завершения
		}
		if value.Done {
			gist = value.Output // окончательный ответ streaming flow
			errI = nil
		}
		return !value.Done // продолжать пока не Done
	})

	for i := range gist {
		log.Debug("get chat gist success",
			slog.Int("batch number", i),
			slog.String("chat gist", gist[i].Gist),
			slog.Int("last message id", gist[i].LastMessageID))
	}

	if errI != nil {
		return nil, errI
	}

	return gist, nil
}

// defineGenerateChatGistFlow определяет сценарий для генерации краткого пересказа чата.
//
//nolint:gocognit
func (s *GenkitService) defineGenerateChatGistFlow() {

	config := &openai.ChatCompletionNewParams{ //конфигурация для OpenRouter provider (OpenAI compatible), для других провайдеров нужно изменять!
		// Основные параметры
		Temperature: openai.Float(0.1), // (0.0 - 2.0) Стабильность (низкие значения) / Креативность (высокие значения)
		// MaxCompletionTokens: openai.Int(1400),  // Максимальное количество токенов в генерируемом ответе (output) Если больше, то ответ будет обрезан.
		// TopP:                openai.Float(0.9), // Ограничивает выбор токенов топ-N% вероятностей. 0.9 — хороший баланс компактности и естественности.
		// Штрафы за повторения
		// FrequencyPenalty: openai.Float(0.8), // Штрафует часто повторяющиеся слова (-2.0 до 2.0). Положительные значения → разнообразие текста.
		// PresencePenalty:  openai.Float(0.6), // Штрафует любые повторяющиеся темы/сущности. 0.2-0.6 → фокус на новых идеях.
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
	/*	prompt := `Тебе дана история сообщений из чата в Telegram.
		Каждое сообщение содержит текст, время отправки, ID отправителя и служебную информацию (например, было ли оно отредактировано, переслано или является ответом).
		Составь КРАТКИЙ и связный пересказ чата (длинной МАКСИМУМ 3500 символов, цель — 3000).
		Сосредоточься на главных событиях, важных решениях, вопросах и ответах.
		Технические детали (например, метки «переслано» или «отредактировано») учитывай только если они влияют на смысл.
		Пересказ должен быть нейтральным, легко читаемым и помогать человеку быстро понять суть обсуждения.
		История сообщений (в хронологическом порядке): {{messages}}
		Приведи только пересказ, без пояснений и дополнительного форматирования.
		`*/
	/*prompt := `Роль: Ты — аналитик чатов, который умеет выделять суть из длинных переписок. Твоя задача — создать структурированный, краткий и информативный пересказ обсуждения. Общий объем пересказа не должен превышать 3900 символов.

	Входные данные:
	Тебе будет предоставлен массив сообщений в формате JSON
	История сообщений (в хронологическом порядке): {{messages}}

	Критически важные инструкции:
	1. Ограничение длины: Итоговый пересказ должен содержать не более 3900 символов. Это абсолютный лимит. Если информации много, агрегируй ее еще сильнее, оставляя только самое главное.
	2. Анализ контекста: Внимательно изучи последовательность сообщений. Учитывай поле reply_to_msg_id, чтобы понимать, какие сообщения являются ответами друг на друга, и восстанавливай логические цепочки.
	3. Строгая фильтрация:
	 - Игнорируй служебные, пустые сообщения, а также стикеры/эмодзи без смысловой нагрузки.
	 - Поля is_edited и is_forwarded упоминай только если это критически важно для понимания (например, если после редактирования изменился смысл решения).
	4. Максимальная агрегация:
	 - Группируй реплики по смыслу, а не перечисляй по отдельности.
	 - Выделяй только ключевые темы, решения, разногласия и действия.
	 - Избегай детализации, цитат и примеров, если они не несут решающей смысловой нагрузки.
	5. Структура итогового пересказа: Сформируй ответ строго в следующем формате:
	 Краткий пересказ чата:
	  - Период: [Начальная дата] - [Конечная дата]
	  - Участники: [Количество уникальных sender_id]
	  - Основные темы: [Список из ключевых тем]
	  - Содержание: [Сплошной, связный текст объемом не более 2-3 коротких абзацев на каждую тему. Здесь изложи всю суть: проблему, обсуждение, аргументы, итог. Будь максимально лаконичным.]
	  - Итог: [Одно-два предложения. Финальный результат или статус обсуждения.]

	Стиль:
	 - Только факты, без вводных слов и домыслов.
	 - Крайне сжатый и информативный стиль.
	 - Хронологический порядок в рамках агрегированного изложения.
	`*/
	prompt := `Роль: Ты — аналитик чатов, который умеет выделять суть из длинных переписок. 
Твоя задача — создать краткий, структурированный и информативный пересказ обсуждения, сгруппированный по ключевым темам. 
Итоговый ответ не должен превышать 3900 символов.

Входные данные:
Тебе будет предоставлен массив сообщений в формате JSON
История сообщений (в хронологическом порядке): {{messages}}

Критически важные инструкции:
1. Ограничение длины: Итоговый пересказ должен содержать не более 3900 символов. Это абсолютный лимит. Если информации много, агрегируй ее еще сильнее, оставляя только самое главное.
2. Фильтрация: Игнорируй служебные сообщения, пустые реплики, стикеры и эмодзи без смысловой нагрузки. Поля is_edited и is_forwarded упоминай только если это критически меняет смысл.
3. Тематический анализ: Выяви от 1 до 3 уникальных ключевых тем обсуждения. Тема — это смысловой кластер сообщений (например, «Планирование встречи», «Обсуждение бюджета», «Решение технической проблемы»).
4. Анонимизация изложения:
 - Запрещено использовать числовые sender_id в тексте пересказа.
 - Вместо этого используй обезличенные формы: «один из участников», «другой участник», «несколько участников», «большинство», «инициатор обсуждения», «критик предложения» и т.п.
 - Если важно указать на разных участников в рамках одной темы, используй минимальные различители: Первый, Второй, Третий участник (но не более 3-х).
5. Структура итогового пересказа: Сформируй ответ строго в следующем формате:
  Краткий пересказ чата:
   - Период: [Дата первого сообщения] — [Дата последнего сообщения]
   - Участники: [количество уникальных sender_id]
   - Основные темы: [Список выявленных тем. Не более 5-и.]
  Содержание (по темам):
  [Здесь представь отдельно по каждой теме из списка «Основные темы»]
  Тема: [Название темы]
  [1 короткий абзац, излагающих суть обсуждения по этой теме. Здесь изложи всю суть: проблему, обсуждение, аргументы, итог. Будь максимально лаконичным. Излагай в хронологическом порядке.]

Общий итог: [2-3 предложения. Резюмируй общий результат или атмосферу всего обсуждения: что в итоге решено, что осталось открытым, какой был настрой участников.]

Стиль:
 - Максимально лаконично, только факты.
 - Внутри темы соблюдай хронологию.
 - Агрегируй информацию: объединяй похожие реплики от одного пользователя, избегай перечисления каждого сообщения.
`

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
				resp, err := s.retryPrompt(ctx, generateChatGistPrompt, batch, log)
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
					Gist:             resp.Text(),
					FirstMessageData: input.Messages[from].Timestamp,
					LastMessageID:    input.Messages[lastMessageID].ID,
					LastMessageData:  input.Messages[lastMessageID].Timestamp,
					MessageCount:     to - from, // кол-во обработанных сообщений в батче, учитывая и пропущенные (пустые, системные и т.п.)
					Audio:            make([]model.AudioGist, 0),
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

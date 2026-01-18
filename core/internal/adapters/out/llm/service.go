// Package llm содержит методы для работы с LLM.
package llm

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	oaic "github.com/firebase/genkit/go/plugins/compat_oai"
	oai "github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/ollama"
)

/*
ENV must be set for specific llm
export OPENROUTER_API_KEY=<your API key>
export GEMINI_API_KEY=<your API key>
export OPENAI_API_KEY=<your API key>
etc...
*/

// GenkitService структура для работы с LLM через фреймворк Genkit.
// Обеспечивает единый интерфейс для взаимодействия с различными LLM-провайдерами
type GenkitService struct {
	g *genkit.Genkit

	contextWindow    int           // context window
	driftPercent     int           // Процент отклонения от заданного контекстного окна (в минус), так как количество токенов можно посчитать только приблизительно.
	symbolPerToken   int           // 1 токен ~ 3 символа. Расчет приблизительный, так как неизвестно как работают токенизаторы различных LLM.
	messagesPerBatch int           // Максимальное количество сообщений в одном запросе к LLM
	flowTimeout      time.Duration // Тайм-аут выполнения сценария LLM

	DefaultTextModel string // Модель для текстового запроса, задается в настройках model дефолтного провайдера (default_provider)

	// TTS
	languageCode             string
	voiceName                string
	currentGeminiApiKeyIndex int // Параметр хранит номер используемого api ключа

	cfg *config.Config

	generateChatGistStreamingFlow *core.Flow[*chat, []model.BatchGist, *int]
	generateAudioGistFlow         *core.Flow[Params, string, struct{}]
}

// withOpenRouter возвращает genkit plugin для работы с платформой агрегатором LLM - OpenRouter.
// OpenAI compatible.
func (s *GenkitService) withOpenRouter() api.Plugin {

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		slog.With("func", "llm.withOpenRouter").Error("OPENROUTER_API_KEY environment variable not set")
		return nil
	}

	openRouterPlugin := &oaic.OpenAICompatible{
		Provider: "openrouter",
		APIKey:   apiKey,
		BaseURL:  "https://openrouter.ai/api/v1",
	}

	/*config := &openai.ChatCompletionNewParams{ //конфигурация
		// Основные параметры
		Temperature:         openai.Float(0.1), // (0.0 - 2.0) Стабильность (низкие значения) / Креативность (высокие значения)
		MaxCompletionTokens: openai.Int(800),   // Максимальное количество токенов в генерируемом ответе (output)
		TopP:                openai.Float(0.9), // Ограничивает выбор токенов топ-N% вероятностей. 0.9 — хороший баланс компактности и естественности.
		// Штрафы за повторения
		FrequencyPenalty: openai.Float(0.8), // Штрафует часто повторяющиеся слова (-2.0 до 2.0). Положительные значения → разнообразие текста.
		PresencePenalty:  openai.Float(0.6), // Штрафует любые повторяющиеся темы/сущности. 0.2-0.6 → фокус на новых идеях.
		// Поведение и формат
		N: openai.Int(1), // Количество вариантов ответа (по умолчанию 1)
		Stop: openai.ChatCompletionNewParamsStopUnion{ // Стоп-сигналы. Например: Stop: openai.F("Пересказ:") — остановка после ключевого слова.
			OfString:      param.Opt[string]{},
			OfStringArray: nil,
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{ // принуждает JSON-ответ.
			OfText:       nil,
			OfJSONSchema: nil,
			OfJSONObject: nil,
		},
	}*/

	return openRouterPlugin
}

// withOllama возвращает genkit plugin для работы с платформой для локального запуска LLM - Ollama.
func (s *GenkitService) withOllama() api.Plugin {

	ollamaPlugin := &ollama.Ollama{
		ServerAddress: s.cfg.LLM.Ollama.ServerAddress,
		Timeout:       s.cfg.LLM.Ollama.Timeout,
	}

	ollamaPlugin.DefineModel(s.g,
		ollama.ModelDefinition{
			Name: s.cfg.LLM.Ollama.Model,
			Type: "chat", // "chat" or "generate"
		},
		&ai.ModelOptions{
			Supports: &ai.ModelSupports{
				Multiturn:  true,
				SystemRole: true,
				Tools:      false,
				Media:      false,
			},
		},
	) // define model

	return ollamaPlugin
}

// withGemini возвращает genkit plugin для работы с семейством моделей (LLM) Gemini AI от Google. nextApiKey указывает на использование следующего api ключа из массива.
func (s *GenkitService) withGemini(nextApiKey bool) api.Plugin {

	log := slog.With("func", "llm.withGemini")

	if nextApiKey {
		s.currentGeminiApiKeyIndex++
		if s.currentGeminiApiKeyIndex == len(s.cfg.LLM.Gemini.ApiKeys) { // Если ключи закончились, начинаем с первого
			s.currentGeminiApiKeyIndex = 0
			log.Info("Gemini Api Keys full rotation")
		}

		log.Info("Set next Gemini API KEY", slog.Int("GeminiApiKeyIndex", s.currentGeminiApiKeyIndex))
	}

	apiKey := s.cfg.LLM.Gemini.ApiKeys[s.currentGeminiApiKeyIndex]

	/*config := &genai.GenerateContentConfig{ // конфигурация
		Temperature: genai.Ptr[float32](1.0), // Устанавливается температура 1.0 — это делает ответы более креативными и менее предсказуемыми.
	}*/

	return &googlegenai.GoogleAI{
		APIKey: apiKey,
	}
}

// withOpenAI возвращает genkit plugin для работы с моделями (LLM) от OpenAI (различные версии GPT).
func (s *GenkitService) withOpenAI() api.Plugin {

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		slog.With("func", "llm.withOpenAI").Error("OPENAI_API_KEY environment variable not set")
		return nil
	}

	oaiPlugin := &oai.OpenAI{
		APIKey: apiKey,
	}

	/*config := &openai.ChatCompletionNewParams{ // Параметры запроса
		Temperature: openai.Float(1.0),
		MaxTokens:   openai.Int(4000),
	}*/

	return oaiPlugin
}

// NewGenkitService инициализация фреймворка для работы с LLM.
// LLM задается через конфигурационный файл.
//
//nolint:gocyclo //cyclo-11
func NewGenkitService(ctx context.Context, cfg *config.Config) (*GenkitService, error) {

	log := slog.With("func", "llm.NewGenkitService")

	if cfg.LLM.Development {
		_ = os.Setenv("GENKIT_ENV", "dev") // During local development (when the `GENKIT_ENV` environment variable is set to `dev`), Init also starts the Reflection API server as a background goroutine. This server provides metadata about registered actions and is used by developer tools. By default, it listens on port 3100.
		log.Info("Start genkit in development mode")
	}

	s := &GenkitService{}

	s.cfg = cfg
	s.flowTimeout = cfg.LLM.FlowTimeout
	s.driftPercent = cfg.LLM.DriftPercent
	s.symbolPerToken = cfg.LLM.SymbolPerToken
	s.messagesPerBatch = cfg.LLM.MessagesPerBatch

	s.languageCode = cfg.LLM.TTS.Gemini.LanguageCode
	s.voiceName = cfg.LLM.TTS.Gemini.VoiceName

	s.initGenkit(ctx, false)

	return s, nil
}

// registerFlows регистрация сценариев (потоков) выполнения промптов.
func (s *GenkitService) registerFlows() {

	s.defineGenerateChatGistFlow()
	s.defineGenerateAudioGistFlow()

}

// initGenkit (ре)инициализация genkit. Если nextApiKey = true, то используется следующий gemini api key из пула.
func (s *GenkitService) initGenkit(ctx context.Context, nextApiKey bool) {

	log := slog.With("func", "llm.initGenkit")

	switch s.cfg.LLM.DefaultProvider {
	case "Ollama":
		s.contextWindow = s.cfg.LLM.Ollama.ContextWindow
		s.DefaultTextModel = "ollama/" + s.cfg.LLM.Ollama.Model
	case "OpenRouter":
		s.contextWindow = s.cfg.LLM.OpenRouter.ContextWindow
		s.DefaultTextModel = "openrouter/" + s.cfg.LLM.OpenRouter.Model
	case "Gemini":
		s.contextWindow = s.cfg.LLM.Gemini.ContextWindow
		s.DefaultTextModel = "googleai/" + s.cfg.LLM.Gemini.Model
	case "OpenAI":
		s.contextWindow = s.cfg.LLM.OpenAI.ContextWindow
		s.DefaultTextModel = "openai/" + s.cfg.LLM.OpenAI.Model
	default:
		log.Error("unknown provider", slog.String("defaultProvider", s.cfg.LLM.DefaultProvider))
	}

	plugins := make([]api.Plugin, 0)
	if s.cfg.LLM.Ollama.Enabled {
		plugins = append(plugins, s.withOllama())
		log.Info("Start Ollama provider")
	}

	if s.cfg.LLM.OpenRouter.Enabled {
		plugins = append(plugins, s.withOpenRouter())
		log.Info("Start OpenRouter provider")
	}

	if s.cfg.LLM.Gemini.Enabled {
		plugins = append(plugins, s.withGemini(nextApiKey))
		log.Info("Start Gemini provider")
	}

	if s.cfg.LLM.OpenAI.Enabled {
		plugins = append(plugins, s.withOpenAI())
		log.Info("Start OpenAI provider")
	}

	// инициализация genkit с заданными в конфигурации плагинами
	s.g = genkit.Init(ctx,
		genkit.WithPlugins(plugins...),
	)

	s.registerFlows()
}

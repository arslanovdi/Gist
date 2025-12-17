package llm

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	oaic "github.com/firebase/genkit/go/plugins/compat_oai"
	oai "github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"google.golang.org/genai"
)

const envFileName = ".env"

// Тип входных данных для запроса к LLM.
type chat struct {
	Messages []model.Message `json:"messages"`
}

type GenkitService struct {
	g      *genkit.Genkit
	config any // Настройки модели, задаются при инициализации фреймворка

	getChatGistFlow *core.Flow[*chat, string, struct{}] // Сценарий (поток) выполнения запросов к LLM
}

// initOpenRouter инициализация genkit для работы с платформой агрегатором LLM - OpenRouter.
// OpenAI compatible.
func (s *GenkitService) initOpenRouter(ctx context.Context, cfg *config.Config) error {

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY environment variable not set")
	}

	openRouterPlugin := &oaic.OpenAICompatible{
		Provider: "openrouter",
		APIKey:   apiKey,
		BaseURL:  "https://openrouter.ai/api/v1",
	}

	s.g = genkit.Init(ctx,
		genkit.WithPlugins(openRouterPlugin),
		genkit.WithDefaultModel("openrouter/"+cfg.LLM.OpenRouter.Model), // Указываем используемую модель (бесплатных моделей было 42 штуки, на момент написания)
	)

	s.config = &openai.ChatCompletionNewParams{ //конфигурация
		Temperature: openai.Float(0.7),
		MaxTokens:   openai.Int(1000),
		TopP:        openai.Float(0.9),
	}

	return nil
}

// initOllama инициализация genkit для работы с платформой для локального запуска LLM - Ollama.
func (s *GenkitService) initOllama(ctx context.Context, cfg *config.Config) {

	ollamaPlugin := &ollama.Ollama{
		ServerAddress: cfg.LLM.Ollama.ServerAddress,
		Timeout:       cfg.LLM.Ollama.Timeout,
	}

	s.g = genkit.Init(ctx,
		genkit.WithPlugins(ollamaPlugin),
		genkit.WithDefaultModel("ollama/"+cfg.LLM.Ollama.Model),
	)

	ollamaPlugin.DefineModel(s.g,
		ollama.ModelDefinition{
			Name: cfg.LLM.Ollama.Model,
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
}

// initGemini инициализация genkit для работы с семейством моделей (LLM) Gemini AI от Google.
func (s *GenkitService) initGemini(ctx context.Context, cfg *config.Config) error {

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	// инициализация genkit с подключенным плагином GoogleAI
	s.g = genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/"+cfg.LLM.Gemini.Model),
	)

	s.config = &genai.GenerateContentConfig{ // конфигурация
		Temperature: genai.Ptr[float32](1.0), // Устанавливается температура 1.0 — это делает ответы более креативными и менее предсказуемыми.
	}

	return nil
}

// initGemini инициализация genkit для работы с моделями (LLM) от OpenAI (различные версии GPT).
func (s *GenkitService) initOpenAI(ctx context.Context, cfg *config.Config) error {

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	oaiPlugin := &oai.OpenAI{
		APIKey: apiKey,
	}

	// инициализация genkit с подключенным плагином OpenAI
	s.g = genkit.Init(ctx,
		genkit.WithPlugins(oaiPlugin),
		genkit.WithDefaultModel("openai/"+cfg.LLM.OpenAI.Model),
	)

	s.config = &openai.ChatCompletionNewParams{ // Параметры запроса
		Temperature: openai.Float(0.5),
		MaxTokens:   openai.Int(100),
	}

	return nil
}

// NewGenkitService инициализация фреймворка для работы с LLM.
// LLM задается через конфигурационный файл.
func NewGenkitService(ctx context.Context, cfg *config.Config) (*GenkitService, error) {

	log := slog.With("func", "llm.NewGenkitService")

	if cfg.LLM.Development {
		_ = os.Setenv("GENKIT_ENV", "dev") // During local development (when the `GENKIT_ENV` environment variable is set to `dev`), Init also starts the Reflection API server as a background goroutine. This server provides metadata about registered actions and is used by developer tools. By default, it listens on port 3100.
		log.Info("Start genkit in development mode")
	}

	// export OPENROUTER_API_KEY=<your API key>
	// export GEMINI_API_KEY=<your API key>
	// export OPENAI_API_KEY=<your API key>
	errE := godotenv.Load(envFileName)
	if errE != nil {
		log.Error("Error loading .env file", slog.Any("error", errE))
	}

	service := &GenkitService{}

	switch cfg.LLM.ClientType {

	case "Ollama":
		service.initOllama(ctx, cfg)
		log.Info("Start Ollama client")

	case "OpenRouter":
		errR := service.initOpenRouter(ctx, cfg)
		if errR != nil {
			return nil, errR
		}
		log.Info("Start OpenRouter client")

	case "Gemini":
		errR := service.initGemini(ctx, cfg)
		if errR != nil {
			return nil, errR
		}
		log.Info("Start Gemini client")

	case "OpenAI":
		errR := service.initOpenAI(ctx, cfg)
		if errR != nil {
			return nil, errR
		}
		log.Info("Start OpenAI client")

	default:
		return nil, fmt.Errorf("unknown client type %s", cfg.LLM.ClientType)
	}

	errF := service.registerFlows()
	if errF != nil {
		return nil, fmt.Errorf("register flows: %w", errF)
	}

	return service, nil
}

// registerFlows регистрация сценариев (потоков) выполнения промптов.
func (s *GenkitService) registerFlows() error {

	s.defineGetChatGistFlow()

	return nil
}

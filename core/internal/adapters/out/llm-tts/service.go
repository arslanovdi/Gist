package llm_tts

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

type GenkitService struct {
	g *genkit.Genkit

	flowTimeout  time.Duration // Тайм-аут выполнения сценария LLM
	languageCode string
	voiceName    string

	generateAudioGistFlow *core.Flow[Params, string, struct{}]
}

// initGemini инициализация genkit для работы с семейством моделей (LLM) Gemini AI от Google.
func (s *GenkitService) initGemini(ctx context.Context, cfg *config.Config) error {

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	s.languageCode = cfg.TTS.Gemini.LanguageCode
	s.voiceName = cfg.TTS.Gemini.VoiceName

	// инициализация genkit с подключенным плагином GoogleAI
	s.g = genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/"+cfg.TTS.Gemini.Model),
	)

	return nil
}

func NewGenkitService(ctx context.Context, cfg *config.Config) (*GenkitService, error) {

	log := slog.With("func", "llm-tts.NewGenkitService")

	if cfg.TTS.Development {
		_ = os.Setenv("GENKIT_ENV", "dev") // During local development (when the `GENKIT_ENV` environment variable is set to `dev`), Init also starts the Reflection API server as a background goroutine. This server provides metadata about registered actions and is used by developer tools. By default, it listens on port 3100.
		log.Info("Start genkit in development mode")
	}

	service := &GenkitService{}
	service.flowTimeout = cfg.TTS.FlowTimeout

	switch cfg.TTS.ClientType {
	case "Gemini":
		errR := service.initGemini(ctx, cfg)
		if errR != nil {
			return nil, errR
		}
		log.Info("Start Gemini-TTS client")
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

	s.defineGenerateAudioGistFlow()

	return nil
}

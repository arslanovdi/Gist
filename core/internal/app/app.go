// Package app инициализация и запуск всех зависимостей, graceful shutdown
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/arslanovdi/Gist/core/internal/adapters/in/tgbot"
	"github.com/arslanovdi/Gist/core/internal/adapters/out/llm"
	"github.com/arslanovdi/Gist/core/internal/adapters/out/tgclient"
	"github.com/arslanovdi/Gist/core/internal/domain/core"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/joho/godotenv"
)

const envFileName = ".env"

// App структура со всеми зависимостями приложения
type App struct {
	Cfg            *config.Config    // Конфигурация
	TelegramBot    *tgbot.Bot        // Телеграм бот
	TelegramClient *tgclient.Session // Телеграм клиент
	CoreService    *core.Gist        // Слой бизнес логики
	LLM            *llm.GenkitService
}

// New создает и инициализирует экземпляр приложения.
// Выполняет настройку всех компонентов в правильном порядке:
//  1. Загружает конфигурацию из .env файла (если доступен)
//  2. Инициализирует Telegram клиент
//  3. Настраивает LLM-сервис
//  4. Создает сервис ядра (бизнес-логика)
//  5. Инициализирует Telegram бота
func New(ctx context.Context) (*App, error) {
	log := slog.With("func", "app.New")

	errE := godotenv.Load(envFileName)
	if errE != nil {
		log.Error("Error loading .env file", slog.Any("error", errE)) // Это корректное поведение, в k8s этого файла может не быть, а параметры передаются через ENV.
	}

	cfg, errC := config.LoadConfig()
	if errC != nil {
		return nil, fmt.Errorf("[app.New] Error loading config: %w", errC)
	}

	log.Info("configuration loaded")

	telegramClient := tgclient.NewSession(cfg)

	llmClient, errL := llm.NewGenkitService(ctx, cfg)
	if errL != nil {
		return nil, fmt.Errorf("[app.new] llm initialization failed: %w", errL)
	}

	/*ttsClient, errT := llm_tts.NewGenkitService(ctx, cfg)
	if errT != nil {
		return nil, fmt.Errorf("[app.new] llm_tts initialization failed: %w", errT)
	}*/
	ttsClient := core.TTSClient(nil)

	coreService := core.NewGist(telegramClient, llmClient, ttsClient, cfg)

	bot, errB := tgbot.New(cfg, coreService)
	if errB != nil {
		return nil, fmt.Errorf("[app.new] bot initialization failed: %w", errB)
	}

	// coreService.SetTelegramBot(bot) // Внедрение зависимости.

	return &App{
		Cfg:            cfg,
		TelegramBot:    bot,
		TelegramClient: telegramClient,
		CoreService:    coreService,
		LLM:            llmClient,
	}, nil
}

// Run запускает приложение и управляет его жизненным циклом.
func (a *App) Run(cancelStartTimeout context.CancelFunc) error {

	log := slog.With("func", "app.Run")

	// Канал ошибок при запуске
	serverErr := make(chan error, 1)
	defer close(serverErr)

	// Запуск всего...
	ctx := context.WithoutCancel(context.Background()) // Нужен долгоживущий контекст (это просто явное его описание).
	a.TelegramClient.Run(ctx, serverErr)
	a.TelegramBot.Run(ctx, serverErr)

	cancelStartTimeout() // все запустили, отменяем контекст запуска приложения
	log.Info("Application started")

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	defer close(stop)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		log.Info("Shutting down...")
	case err := <-serverErr:
		log.Info("Error on Run. Shutdown...", slog.Any("error", err))
	}

	ctxShutdown, cancelShutdownTimeout := context.WithTimeout(context.Background(), a.Cfg.Project.ShutdownTimeout) // контекст останова приложения
	defer cancelShutdownTimeout()
	go func() {
		<-ctxShutdown.Done()
		if errors.Is(ctxShutdown.Err(), context.DeadlineExceeded) {
			log.Warn("Application shutdown time exceeded")
		}
	}()

	a.Close(ctxShutdown)

	return nil
}

// Close операции по остановке работы с зависимостями.
func (a *App) Close(ctx context.Context) {

	log := slog.With("func", "app.Close")

	a.TelegramBot.Close(ctx)
	a.TelegramClient.Close(ctx)

	log.Info("Application stopped")
}

package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/arslanovdi/Gist/core/internal/tgbot"
	"github.com/arslanovdi/Gist/core/internal/user"
	"github.com/joho/godotenv"
)

type App struct {
	Cfg *config.Config
	Bot *tgbot.Bot
}

func New() (*App, error) {
	log := slog.With("func", "app.New")

	errE := godotenv.Load(".env")
	if errE != nil {
		log.Error("Error loading .env file", slog.Any("error", errE)) // Это корректное поведение, т.к. в k8s этого файла может не быть, а параметры передаются через ENV. Просто логируем поведение.
	}

	cfg, errC := config.LoadConfig()
	if errC != nil {
		return nil, fmt.Errorf("[app.New] Error loading config: %w", errC)
	}

	log.Info("configuration loaded")

	userCommands := user.New(cfg)

	bot, errB := tgbot.New(cfg, userCommands)
	if errB != nil {
		return nil, fmt.Errorf("[app.new] bot initialization failed: %w", errB)
	}

	return &App{
		Cfg: cfg,
		Bot: bot,
	}, nil
}

func (a *App) Run(cancelStartTimeout context.CancelFunc) error {

	log := slog.With("func", "app.Run")

	// Канал ошибок при запуске
	serverErr := make(chan error, 1)
	defer close(serverErr)

	// TODO Запуск всего...
	ctx := context.WithoutCancel(context.Background())
	a.Bot.Run(ctx, serverErr)

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

func (a *App) Close(ctx context.Context) {

	log := slog.With("func", "app.Close")

	a.Bot.Close(ctx)

	log.Info("Application stopped")
}

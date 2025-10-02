package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
)

type App struct {
	Config *config.Config
}

func New() *App {
	log := slog.With("func", "app.New")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("configuration loaded")

	return &App{
		Config: cfg,
	}
}

// Run
// Инициализация зависимостей, запуск бота
func (a *App) Run(_ context.Context, cancelStart context.CancelFunc) error {

	log := slog.With("func", "app.Run")

	// TODO Запуск всего...

	stop := make(chan os.Signal, 1)
	defer close(stop)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	cancelStart() // все запустили, отменяем контекст запуска приложения

	select {
	case <-stop:
		log.Info("Shutting down...")
	}

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), a.Config.Project.ShutdownTimeout) // контекст останова приложения
	defer cancelShutdown()
	go func() {
		<-ctxShutdown.Done()
		if errors.Is(ctxShutdown.Err(), context.DeadlineExceeded) {
			log.Warn("Application shutdown time exceeded")
		}
	}()

	a.Close(ctxShutdown)

	return nil
}

func (a *App) Close(_ context.Context) {

	log := slog.With("func", "app.Close")

	log.Info("Application stopped")
}

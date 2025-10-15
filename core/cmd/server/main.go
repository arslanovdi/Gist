package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/arslanovdi/Gist/core/internal/app"
	"github.com/arslanovdi/Gist/core/internal/infra/logger"
)

const (
	startupTimeout = 60 // seconds
	serviceName    = "core"
)

func main() {
	// контекст запуска приложения
	ctxStart, cancelStartTimeout := context.WithTimeout(context.Background(), time.Duration(startupTimeout)*time.Second)
	defer cancelStartTimeout()
	go func() {
		<-ctxStart.Done()
		if errors.Is(ctxStart.Err(), context.DeadlineExceeded) { // приложение зависло при запуске
			slog.With("func", "main").Error("Application startup time exceeded")
			os.Exit(1)
		}
	}()

	logger.InitializeLogger(slog.LevelDebug, serviceName)
	log := slog.With("func", "main")

	application, errA := app.New()
	if errA != nil {
		log.Error("Failed to initialize application", slog.Any("error", errA))
		os.Exit(1)
	}

	if application.Cfg.Project.Debug {
		logger.SetLogLevel(slog.LevelDebug)
	} else {
		logger.SetLogLevel(slog.LevelInfo)
	}

	if err := application.Run(cancelStartTimeout); err != nil {
		log.Error("failed to run application", slog.Any("error", err))
	}
}

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(startupTimeout)*time.Second)
	defer cancel()
	go func() {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) { // приложение зависло при запуске
			slog.Error("Application startup time exceeded")
			os.Exit(1)
		}
	}()

	logger.InitializeLogger(slog.LevelDebug, serviceName)

	application := app.New()

	if application.Config.Project.Debug {
		logger.SetLogLevel(slog.LevelDebug)
	} else {
		logger.SetLogLevel(slog.LevelInfo)
	}

	if err := application.Run(ctx, cancel); err != nil {
		slog.With("function", "main").Error("failed to run application", slog.Any("error", err))
	}
}

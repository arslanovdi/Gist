// Package logger содержит методы для работы с логгированием.
package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

var (
	options     *slog.HandlerOptions
	loglevel    *slog.LevelVar
	serviceName string
)

// contextHandler - обертка над стандартным slog.Handler, которая добавляет
// информацию о трассировке (trace_id и span_id) в логи, если они присутствуют в контексте и сэмплированы.
type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}

// Handle adds trace_id, span_id attributes to log from context
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {

	if r.Level == slog.LevelDebug {
		switch r.Message {
		case "Action.Run":
			return nil // Отключаем логирование genkit message Action.Run, он туда пишет все контекстное окно и забивает логи сообщениями по 500_000 символов...
		}
	}

	span := trace.SpanFromContext(ctx)

	sCtx := span.SpanContext()

	if sCtx.IsValid() && sCtx.IsSampled() {
		r.Add("trace_id", sCtx.TraceID().String())
		r.Add("span_id", sCtx.SpanID().String())
	}

	return h.Handler.Handle(ctx, r)
}

// InitializeLogger initializes the slog logger
func InitializeLogger(level slog.Level, service string) {
	hidePassword := func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == "password" {
			return slog.String("password", "********")
		}
		return a
	}
	loglevel = &slog.LevelVar{}
	loglevel.Set(level)

	options = &slog.HandlerOptions{
		AddSource:   false,
		ReplaceAttr: hidePassword,
		Level:       loglevel,
	}

	serviceName = service

	baseHandler := slog.NewJSONHandler(os.Stderr, options)
	wrappedHandler := &contextHandler{baseHandler}

	logger := slog.New(wrappedHandler).With("service", serviceName)

	// В проекте используется глобальный логгер
	slog.SetDefault(logger)

	slog.With("func", "logger.InitializeLogger").Info("Logger initialized", slog.String("level", loglevel.Level().String()))
}

// SetLogLevel sets the level of the logger
// В пакете slog нет метода установки уровня логирования для глобального логгера, создаем свой
func SetLogLevel(level slog.Level) {
	log := slog.With("func", "logger.SetLogLevel")

	if options == nil {
		InitializeLogger(slog.LevelDebug, serviceName)
	}

	loglevel.Set(level)
	log.Info("Set logger level", slog.String("level", level.Level().String()))
}

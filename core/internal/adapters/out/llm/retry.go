package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
)

// Retry логика с экспоненциальным backoff для 429
func retryPrompt(ctx context.Context, prompt ai.Prompt, input any, log *slog.Logger) (*ai.ModelResponse, error) {

	maxRetries := 7
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := prompt.Execute(ctx, ai.WithInput(input))
		if err == nil {
			return resp, nil // Успех
		}

		// Проверяем, ретраить ли данную ошибку от OpenRouter TODO ошибки других LLM провайдеров пока не ловил.
		if isErrorToRetry(err) {
			if attempt == maxRetries {
				return nil, fmt.Errorf("max retries exceeded for rate limit: %w", err)
			}

			delay := baseDelay * time.Duration(1<<attempt) // 1s, 2s, 4s, 8s, 16s

			log.Warn("Rate limit hit, retrying",
				slog.Int("attempt", attempt+1),
				slog.Int("max_retries", maxRetries),
				slog.Duration("delay", delay),
				slog.String("error", err.Error()))

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			continue
		}

		// Для других ошибок - не ретраим
		return nil, fmt.Errorf("prompt execute failed: %w", err)
	}
	return nil, fmt.Errorf("unreachable")
}

// Проверяет, является ли ошибка 429 или 502 от OpenRouter
func isErrorToRetry(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "429") || // 429 Too Many Requests
		strings.Contains(errStr, "502") // 502 Bad Gateway
}

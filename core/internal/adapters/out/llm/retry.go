package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/firebase/genkit/go/ai"
	"google.golang.org/genai"
)

// llm от Xiaomi, зафиксированный тайм-аут ответа 11.40m !!!

// Retry логика с экспоненциальным backoff для ошибок. // TODO и тайм-аутом ответа от llm в 60 секунд.
// Если с ошибкой прилетает время задержки, то выбирается оно, вместо экспоненциального.
func (s *GenkitService) retryPrompt(ctx context.Context, prompt ai.Prompt, input any, log *slog.Logger) (*ai.ModelResponse, error) {

	start := time.Now()

	maxRetries := 7
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		ctxPrompt, cancelPrompt := context.WithTimeout(ctx, s.cfg.LLM.PromptTimeout)
		log.Debug("Запуск промпта", slog.Int("попытка", attempt))
		resp, err := prompt.Execute(ctxPrompt, ai.WithInput(input))
		if err == nil {
			log.Debug("запрос к llm выполнен успешно", slog.Any("время обработки", time.Since(start).String()))
			cancelPrompt()
			return resp, nil // Успех
		}
		cancelPrompt()

		// Проверяем, ретраить ли данную ошибку
		if isErrorToRetry(err) { // TODO может быть ретраить ВСЕ ошибки?
			if attempt == maxRetries {
				return nil, fmt.Errorf("max retries exceeded for rate limit: %w", err)
			}

			delay := baseDelay * time.Duration(1<<attempt) // 1s, 2s, 4s, 8s, 16s

			genaiErr := genai.APIError{}
			if errors.As(err, &genaiErr) {

				if genaiErr.Status == "RESOURCE_EXHAUSTED" {
					log.Warn("genai error", slog.String("error", genaiErr.Error()))
					return nil, model.ErrResourceExhausted
				}

				if len(genaiErr.Details) >= 2 {
					if retryDelayVal, ok := genaiErr.Details[2]["retryDelay"]; ok { // Во время отладки ловил значение в 3-м слайсе. Если структура изменится это может стать проблемой.
						if retryDelayStr, ok := retryDelayVal.(string); ok {
							parsedDelay, errP := time.ParseDuration(retryDelayStr)
							if errP == nil {
								delay = parsedDelay // Время задержки полученное из ошибки

								log.Info("retry delay from genai error", slog.Any("delay", delay.String()))
							}
						}
					}
				}
			}

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

		log.Debug("запрос к llm НЕ выполнен", slog.Any("время обработки", time.Since(start).String()))
		// Для других ошибок - не ретраим
		return nil, fmt.Errorf("prompt execute failed: %w", err)
	}
	return nil, fmt.Errorf("unreachable")
}

// Проверяет, является ли ошибка 429, 502, 503 и еще несколько.
func isErrorToRetry(err error) bool {
	errStr := strings.ToLower(err.Error())
	return errors.Is(err, context.DeadlineExceeded) || // Дедлайн промпта
		strings.Contains(errStr, "429") || // 429 Too Many Requests	(OpenRouter, Ollama, Gemini)
		strings.Contains(errStr, "502") || // 502 Bad Gateway	(OpenRouter)
		strings.Contains(errStr, "503") || // 503 The model is overloaded	(Gemini)
		strings.Contains(errStr, "no choices in completion") // Ошибка возвращается если модель перегружена, бывает у free моделей (OpenRouter)
}

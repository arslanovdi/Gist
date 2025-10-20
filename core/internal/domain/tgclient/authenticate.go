package tgclient

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// Authenticate ожидание номера телефона и кода подтверждения, авторизация телеграм клиента.
//
// Параметры:
//
//	ctx - контекст должен быть с таймаутом, иначе могут плодиться горутины.
//
// Возвращает error типа ErrUnauthorized, с каналами в виде payload.
func (s *Session) Authenticate(ctx context.Context) error {
	log := slog.With("func", "user.authenticate", slog.Any("user_id", s.UserID))

	s.Cred.Phone = ""

	phone := make(chan string)
	code := make(chan string)
	authError := make(chan error)
	wgCred := &sync.WaitGroup{}

	phoneDone := make(chan struct{})

	wgCred.Go(func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				log.Debug("context canceled", slog.Any("error", ctx.Err()))
				break loop
			case s.Cred.Phone = <-phone:
				log.Debug("getting credential", slog.Any("phone", s.Cred.Phone))
				close(phoneDone)
			case s.Cred.Code = <-code:
				log.Debug("getting credential", slog.Any("code", s.Cred.Code))
				break loop
			}
		}

		close(phone) // запись в закрытый канал невозможна, так как у этого метода и метода пишущего в канал - общий контекст.
		close(code)
		log.Debug("get credentials done")
	})

	go func() {
		if err := s.Client.Run(ctx, func(ctx context.Context) error {
			// Функция для запроса кода подтверждения
			codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
				log.Debug("Ожидаем credentials")
				wgCred.Wait() // ждем пока код прилетит по каналу и сохранится в поле Cred.
				log.Debug("code prompt done")
				return s.Cred.Code, nil
			}

			<-phoneDone // Ждем получения номера телефона!

			// Аутентификация без пароля
			flow := auth.NewFlow(
				auth.CodeOnly(s.Cred.Phone, auth.CodeAuthenticatorFunc(codePrompt)),
				auth.SendCodeOptions{},
			)

			errF := flow.Run(ctx, s.Client.Auth())
			if errF != nil {
				// При ошибке аутентификации удаляем сессию и пробуем снова
				if errR := os.Remove("session.json"); errR != nil && !os.IsNotExist(errR) {
					log.Error("Warning: failed to remove session file", slog.Any("error", errR))
				}
				log.Debug("authentication failed", slog.Any("error", errF), slog.Int64("user_id", s.UserID))
				return fmt.Errorf("authentication failed: %w", errF)
			}

			log.Debug("Authentication successful!", slog.Int64("user_id", s.UserID))

			return nil
		}); err != nil {
			authError <- err
		} else {
			authError <- nil // Успешная аутентификация без пароля
		}

		close(authError)
	}()

	return model.NewErrUnauthorized(phone, code, authError) // Возвращаем ошибку с каналами.
}

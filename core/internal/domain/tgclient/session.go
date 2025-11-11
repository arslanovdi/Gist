package tgclient

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/gotd/td/telegram"
)

const batchLimit = 100

type Session struct {
	userID int64  // Идентификатор пользователя Telegram
	phone  string // Номер телефона, привязанный к аккаунту

	client     *telegram.Client
	wg         *sync.WaitGroup
	ready      atomic.Bool        // True - клиент готов к работе
	cancelFunc context.CancelFunc // Отмена контекста вызовет закрытие telegram.Client.
}

func NewSession(cfg *config.Config) *Session {
	// Настройка клиента Telegram с сохранением сессии
	client := telegram.NewClient(
		cfg.Client.AppID,
		cfg.Client.AppHash,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{ // TODO Реализовать сохранение во внешнее хранилище сессий
				Path: "session.json",
			},
		},
	)

	return &Session{
		userID: cfg.Client.UserID,
		phone:  cfg.Client.Phone,
		client: client,
		wg:     &sync.WaitGroup{},
	}
}

func (s *Session) Run(ctx context.Context, serverErr chan error) {
	log := slog.With("func", "tgclient.Run")
	log.Debug("run telegram client")

	if serverErr == nil {
		log.Error("error channel is nil")
		return
	}

	ctxClient, cancelClient := context.WithCancel(ctx)
	s.cancelFunc = cancelClient

	// Запускаем клиент в отдельной горутине
	s.wg.Go(func() {
		log.Debug("Starting Telegram client...")

		// Запуск клиента
		if err := s.client.Run(ctxClient, func(ctx context.Context) error {
			// Проверяем, авторизованы ли мы уже
			authStatus, err := s.client.Auth().Status(ctx)
			if err != nil {
				return fmt.Errorf("get auth status failed: %w", err)
			}

			// Если не авторизованы, выполняем полный процесс авторизации
			if !authStatus.Authorized {
				log.Debug("Not authenticated, starting authentication flow...", slog.Int64("user_id", s.userID))
				if errA := s.Authenticate(ctx); errA != nil {
					return errA
				}
			} else {
				log.Debug("Already authenticated, using existing session...", slog.Int64("user_id", s.userID))
			}

			// Сигнализируем, что клиент готов
			s.ready.Store(true)
			log.Debug("Telegram client is ready")

			// Ждем отмены контекста
			<-ctx.Done()
			s.ready.Store(false)
			return nil
		}); err != nil {
			log.Error("Telegram client stopped with error", slog.Any("error", err))
			s.ready.Store(false)
			serverErr <- err
		}
	})
}

func (s *Session) Close(ctx context.Context) {
	log := slog.With("func", "tgclient.Stop")
	log.Debug("Stopping Telegram client...")

	if s.cancelFunc != nil {
		s.cancelFunc()
	} else {
		log.Error("Telegram client has not been started")
	}

	done := make(chan struct{})
	go func() { // Ожидаем завершения горутин в отдельном потоке.
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Debug("Telegram client stopped")
	case <-ctx.Done():
		log.Error("Context canceled before Telegram client stopped", slog.Any("error", ctx.Err()))
	}
}

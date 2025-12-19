// Package tgclient реализация телеграмм клиента
package tgclient

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"golang.org/x/time/rate"
)

const batchLimit = 100

// Session структура телеграм клиента. В этой реализации подключение только одного пользователя!
type Session struct {
	userID int64  // Идентификатор пользователя Telegram
	phone  string // Номер телефона, привязанный к аккаунту

	client     *telegram.Client
	wg         *sync.WaitGroup
	ready      atomic.Bool        // True - клиент готов к работе
	cancelFunc context.CancelFunc // Отмена контекста вызовет закрытие telegram.Client.
	waiter     *floodwait.Waiter
}

// NewSession создает и инициализирует новый экземпляр сессии Telegram клиента.
// Принимает конфигурацию приложения, содержащую:
//   - AppID и AppHash для аутентификации в Telegram API
//   - UserID и Phone пользователя
//
// Возвращает готовый к использованию экземпляр Session.
// Сессия сохраняется локально в файл "session.json".
func NewSession(cfg *config.Config) *Session {

	// обработчик ошибки FlOOD_WAIT
	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		slog.Error("Got FLOOD_WAIT", slog.Any("sleep", wait.Duration))
	})

	// Настройка клиента Telegram с сохранением сессии
	client := telegram.NewClient(
		cfg.Client.AppID,
		cfg.Client.AppHash,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{ // TODO Реализовать сохранение во внешнее хранилище сессий
				Path: "session.json",
			},
			Middlewares: []telegram.Middleware{
				waiter, // обработчик FLOOD_WAIT
				ratelimit.New(rate.Every(100*time.Millisecond), 5), // Общий rate limit, чтобы реже ловить FLOOD_WAIT. Субъективно, не особо помогает.
			},
		},
	)

	return &Session{
		userID: cfg.Client.UserID,
		phone:  cfg.Client.Phone,
		client: client,
		wg:     &sync.WaitGroup{},
		waiter: waiter,
	}
}

// Run запускает клиент Telegram в отдельной горутине.
// Управляет жизненным циклом клиента, включая аутентификацию и обработку ошибок.
// Принимает:
//   - ctx: контекст для управления временем жизни клиента
//   - serverErr: канал для отправки критических ошибок в родительскую горутину
//
// Автоматически:
//   - Проверяет аутентификацию
//   - Запускает процесс аутентификации при необходимости
//   - Устанавливает флаг готовности
//   - Обрабатывает завершение работы
//
// При возникновении ошибки отправляет её в канал serverErr.
//
//nolint:gocognit //cognit-14
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
			return s.waiter.Run(ctx, func(ctx context.Context) error { // Оборачиваем клиентский код в waiter
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
			})
		}); err != nil {
			log.Error("Telegram client stopped with error", slog.Any("error", err))
			s.ready.Store(false)
			serverErr <- err
		}
	})
}

// Close корректно завершает работу клиента Telegram.
// Останавливает все запущенные горутины и освобождает ресурсы.
// Принимает контекст для контроля времени ожидания завершения.
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

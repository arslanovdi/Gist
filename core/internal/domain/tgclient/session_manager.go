package tgclient

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/gotd/td/telegram"
)

type SessionManager struct {
	appID   int
	appHash string
	ttl     time.Duration // 30 минут по умолчанию
	close   chan struct{}

	userSessions sync.Map // map[int64]*user.Session	TODO Это некое состояние. Обязательно ли хранить клиентскую сессию в памяти? Или создавать новую при каждом запросе (это увеличит нагрузку на хранилище сессий).
}

func NewSessionManager(cfg *config.Config) *SessionManager {

	manager := &SessionManager{
		appID:        cfg.Bot.AppID,
		appHash:      cfg.Bot.AppHash,
		ttl:          cfg.TgClient.SessionTTL,
		userSessions: sync.Map{},
		close:        make(chan struct{}),
	}

	go func() { // cron Очистка устаревших сессий
		for {
			select {
			case <-time.After(time.Minute):
				manager.CleanupExpiredSessions()
			case <-manager.close:
				return
			}
		}
	}()

	return manager
}

func (m *SessionManager) GetSession(ctx context.Context, userID int64) (*Session, error) {
	log := slog.With("func", "user.getSession", "user_id", userID)

	// Попытка найти существующую сессию
	if session, exist := m.userSessions.Load(userID); exist {
		log.Debug("got session from cache")
		userSession := session.(*Session)
		userSession.UpdateLastAccess()

		return userSession, nil
	}

	session, err := m.CreateSession(userID)
	if err != nil {
		return nil, err
	}

	return nil, session.Authenticate(ctx)
}

func (m *SessionManager) CreateSession(userID int64) (*Session, error) {
	log := slog.With("func", "user.CreateSession", "user_id", userID)

	// Настройка клиента Telegram с сохранением сессии
	client := telegram.NewClient(
		m.appID,
		m.appHash,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{ // TODO Реализовать сохранение во внешнее хранилище сессий
				Path: "session.json", // TODO только на время разработки
			},
		},
	)

	session := &Session{
		UserID:     userID,
		Client:     client,
		LastAccess: time.Now(),
		CreatedAt:  time.Now(),
	}

	m.userSessions.Store(userID, session)

	log.Debug("created session")
	return session, nil
}

func (m *SessionManager) CloseSession(userID int64) {
	if _, exists := m.userSessions.Load(userID); exists {
		m.userSessions.Delete(userID)
	}
}

func (m *SessionManager) CleanupExpiredSessions() {
	m.userSessions.Range(func(key, value any) bool {
		userID := key.(int64)
		session := value.(*Session)

		if time.Since(session.LastAccess) > m.ttl {
			m.CloseSession(userID)
		}
		return true
	})
}

func (m *SessionManager) Close() {
	close(m.close)
}

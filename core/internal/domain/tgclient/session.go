package tgclient

import (
	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/gotd/td/telegram"
)

type Session struct {
	UserID int64
	Phone  string
	Client *telegram.Client
}

func NewSession(cfg *config.Config) *Session {
	// Настройка клиента Telegram с сохранением сессии
	client := telegram.NewClient(
		cfg.Client.AppID,
		cfg.Client.AppHash,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{ // TODO Реализовать сохранение во внешнее хранилище сессий
				Path: "session.json", // TODO только на время разработки
			},
		},
	)

	return &Session{
		UserID: cfg.Client.UserID,
		Phone:  cfg.Client.Phone,
		Client: client,
	}
}

package core

import (
	"context"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
)

type TelegramClient interface {
	GetAllChats(ctx context.Context) ([]model.Chat, error)
}

type TelegramBot interface {
	Authentication(ctx context.Context, userID int64, phone, code chan<- string, authError chan error) error // Запрос credentials у пользователя
}

// TODO Политика реагирования на отказ авторизации клиента.
// Запрашиваются credentials и проходит авторизация.

type Gist struct {
	tgClient TelegramClient
	tgBot    TelegramBot

	UnreadThreshold int

	requestTimeout time.Duration
}

func NewGist(tgClient TelegramClient, cfg *config.Config) *Gist {
	return &Gist{
		tgClient:        tgClient,
		requestTimeout:  cfg.Client.RequestTimeout,
		UnreadThreshold: cfg.Settings.ChatUnreadThreshold,
	}
}

// SetTelegramBot внедрение зависимости, т.к. бот и слой бизнес-логики должны вызывать методы друг-друга.
func (g *Gist) SetTelegramBot(tgBot TelegramBot) {
	g.tgBot = tgBot
}

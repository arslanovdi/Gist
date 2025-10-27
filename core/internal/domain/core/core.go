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

// Gist представляет основной сервис приложения, который инкапсулирует бизнес-логику.
type Gist struct {
	tgClient TelegramClient

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

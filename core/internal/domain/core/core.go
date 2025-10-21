package core

import (
	"context"
	"log/slog"
	"time"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
)

type TelegramClient interface {
	GetAllChats(ctx context.Context) ([]string, error)
}

type TelegramBot interface {
	Authentication(ctx context.Context, userID int64, phone, code chan<- string, authError chan error) error // Запрос credentials у пользователя
}

// TODO Политика реагирования на отказ авторизации клиента.
// Текущий запрос не выполнится, запрашиваются credentials и проходит авторизация.
// Блок проверки авторизации необходимо вставлять во все методы, вызывающие tgClient, т.к. она может протухнуть в любой момент.

type Gist struct {
	tgClient TelegramClient
	tgBot    TelegramBot

	requestTimeout time.Duration
}

func (g *Gist) GetAllChats(_ context.Context) ([]string, error) {
	log := slog.With("func", "core.GetAllChats")

	log.Debug("Get all chats")

	ctxClient, cancelClient := context.WithTimeout(context.TODO(), g.requestTimeout) // Контекст ограничивающий время выполнения запроса (включая закрытие горутин аутентификации в боте и клиенте по таймауту)
	defer cancelClient()

	chats, errG := g.tgClient.GetAllChats(ctxClient)
	if errG != nil {
		return nil, errG // Прочие ошибки
	}

	log.Debug("Successfully get all chats", slog.Any("chats", chats))

	return chats, nil
}

func NewGist(tgClient TelegramClient, cfg *config.Config) *Gist {
	return &Gist{
		tgClient:       tgClient,
		requestTimeout: cfg.Client.RequestTimeout,
	}
}

// SetTelegramBot внедрение зависимости, т.к. бот и слой бизнес-логики должны вызывать методы друг-друга.
func (g *Gist) SetTelegramBot(tgBot TelegramBot) {
	g.tgBot = tgBot
}

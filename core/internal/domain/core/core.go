package core

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
)

type TelegramClient interface {
	GetAllChats(ctx context.Context, userID int64) ([]string, error)
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

func (g *Gist) GetAllChats(_ context.Context, userID int64) ([]string, error) {
	log := slog.With("func", "core.GetAllChats")

	log.Debug("Get all chats", slog.Int64("user_id", userID))

	ctxClient, cancelClient := context.WithTimeout(context.TODO(), g.requestTimeout) // Контекст ограничивающий время выполнения запроса (включая закрытие горутин аутентификации в боте и клиенте по таймауту)
	defer cancelClient()

	chats, errG := g.tgClient.GetAllChats(ctxClient, userID)
	if errG != nil {
		var unauthorizedErr *model.ErrUnauthorized
		if errors.As(errG, &unauthorizedErr) { // Ошибка авторизации.

			log.Debug("unauthorized telegram client", slog.Int64("user_id", userID), slog.Any("err", errG))

			errA := g.tgBot.Authentication(ctxClient, userID, unauthorizedErr.Phone, unauthorizedErr.Code, unauthorizedErr.AuthError) // Запрос credentials у пользователя
			if errA != nil {                                                                                                          // Ошибка получения credentials
				return nil, errA
			}
		} else {
			return nil, errG // Прочие ошибки
		}
	}

	log.Debug("Successfully get all chats", slog.Int64("user_id", userID), slog.Any("chats", chats))

	return chats, nil
}

func NewGist(tgClient TelegramClient, cfg *config.Config) *Gist {
	return &Gist{
		tgClient:       tgClient,
		requestTimeout: cfg.TgClient.RequestTimeout,
	}
}

// SetTelegramBot внедрение зависимости, т.к. бот и слой бизнес-логики должны вызывать методы друг-друга.
func (g *Gist) SetTelegramBot(tgBot TelegramBot) {
	g.tgBot = tgBot
}

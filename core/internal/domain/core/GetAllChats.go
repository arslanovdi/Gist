package core

import (
	"context"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// GetAllChats возвращает список всех чатов пользователя.
func (g *Gist) GetAllChats(ctx context.Context) ([]model.Chat, error) {
	log := slog.With("func", "core.GetAllChats")

	log.Debug("Get all chats")

	ctxClient, cancelClient := context.WithTimeout(ctx, g.requestTimeout) // Контекст ограничивающий время выполнения запроса (включая закрытие горутин аутентификации в боте и клиенте по таймауту)
	defer cancelClient()

	chats, errG := g.tgClient.GetAllChats(ctxClient)
	if errG != nil {
		return nil, errG // Прочие ошибки
	}

	log.Debug("Successfully get all chats", slog.Any("chats", chats))

	return chats, nil
}

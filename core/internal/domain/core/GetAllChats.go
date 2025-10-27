package core

import (
	"context"
	"log/slog"
	"sort"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

func (g *Gist) GetAllChats(ctx context.Context) ([]model.Chat, error) {
	log := slog.With("func", "core.GetAllChats")

	log.Debug("Get all chats")

	ctxClient, cancelClient := context.WithTimeout(ctx, g.requestTimeout) // Контекст ограничивающий время выполнения запроса (включая закрытие горутин аутентификации в боте и клиенте по таймауту)
	defer cancelClient()

	chats, errG := g.tgClient.GetAllChats(ctxClient)
	if errG != nil {
		return nil, errG // Прочие ошибки
	}

	// отсортировать по убыванию UnreadCount
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].UnreadCount > chats[j].UnreadCount
	})

	log.Debug("Successfully get all chats", slog.Any("chats", chats))

	return chats, nil
}

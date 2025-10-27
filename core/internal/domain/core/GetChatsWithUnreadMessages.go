package core

import (
	"context"
	"log/slog"
	"sort"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

func (g *Gist) GetChatsWithUnreadMessages(ctx context.Context) ([]model.Chat, error) {

	log := slog.With("func", "core.GetChatsWithUnreadMessages")
	log.Debug("get chats with unread messages")

	ctxClient, cancelClient := context.WithTimeout(ctx, g.requestTimeout) // Контекст ограничивающий время выполнения запроса
	defer cancelClient()

	chats, errG := g.tgClient.GetAllChats(ctxClient)
	if errG != nil {
		return nil, errG // Прочие ошибки
	}

	// отсортировать по убыванию UnreadCount
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].UnreadCount > chats[j].UnreadCount
	})

	// Отбрасываем чаты, где UnreadCount < UnreadThreshold
	if len(chats) == 0 {
		return nil, nil
	}
	i := 0
	for chats[i].UnreadCount >= g.UnreadThreshold {
		i++
	}
	undreadChats := make([]model.Chat, i)
	copy(undreadChats, chats[:i])

	log.Debug("Successfully get all chats", slog.Any("chats", chats))

	return undreadChats, nil

}

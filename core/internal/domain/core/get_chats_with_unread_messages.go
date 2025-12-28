package core

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// GetChatsWithUnreadMessages возвращает список чатов с непрочитанными сообщениями.
//
// Отбирает чаты, где количество непрочитанных сообщений больше или равно пороговому значению UnreadThreshold
func (g *Gist) GetChatsWithUnreadMessages(ctx context.Context) ([]model.Chat, error) {
	log := slog.With("func", "core.GetChatsWithUnreadMessages")
	log.Debug("get chats with unread messages")

	chats, errA := g.GetAllChats(ctx)
	if errA != nil {
		return nil, fmt.Errorf("GetChatsWithUnreadMessages: %w", errA)
	}

	unreadChats := make([]model.Chat, 0)
	// Отбрасываем чаты, где UnreadCount < UnreadThreshold
	if len(chats) == 0 {
		return nil, fmt.Errorf("GetChatsWithUnreadMessages: no chat found") // TODO обработать ошибку выше
	}
	for i := range chats {
		if chats[i].UnreadCount >= g.UnreadThreshold {
			unreadChats = append(unreadChats, chats[i])
		}
	}
	// отсортировать по убыванию UnreadCount
	sort.Slice(unreadChats, func(i, j int) bool {
		return unreadChats[i].UnreadCount > unreadChats[j].UnreadCount
	})

	log.Debug("Successfully get chats with unread messages", slog.Any("unread chats count", len(unreadChats)))

	return unreadChats, nil

}

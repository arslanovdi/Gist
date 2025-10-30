package core

import (
	"context"
	"fmt"
	"log/slog"

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

	// Отбрасываем чаты, где UnreadCount < UnreadThreshold
	if len(chats) == 0 {
		return nil, nil
	}
	i := 0
	for chats[i].UnreadCount >= g.UnreadThreshold {
		i++
	}
	unreadChats := make([]model.Chat, i)
	copy(unreadChats, chats[:i])

	log.Debug("Successfully get chats with unread messages", slog.Any("unreadChats", unreadChats))

	return unreadChats, nil

}

package core

import (
	"context"
	"fmt"
	"slices"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// MarkAsRead отметить сообщения чата как прочитанные. Нумерация страниц начинается с 1.
func (g *Gist) MarkAsRead(ctx context.Context, chatID int64, pageID int) (*model.Chat, error) {

	chat, errD := g.GetChatDetail(ctx, chatID)
	if errD != nil {
		return nil, fmt.Errorf("core.MarkAsRead: %w", errD)
	}

	lastMessageID := 0 // Если страница не задана == 0, отмечаем прочитанными ВСЕ сообщения чата.
	if pageID > 0 {    // Иначе отмечаем прочитанными только сообщения до текущего батча с кратким пересказом.
		lastMessageID = chat.Gist[pageID-1].LastMessageID
	}

	errM := g.tgClient.MarkAsRead(ctx, chat, lastMessageID)
	if errM != nil {
		return nil, fmt.Errorf("core.MarkAsRead: %w", errM)
	}

	if pageID > 0 {
		chat.UnreadCount -= chat.Gist[pageID-1].MessageCount // уменьшаем количество непрочитанных сообщений в чате
		chat.Gist = slices.Delete(chat.Gist, pageID-1, 1)
		return chat, nil
	}

	chat.UnreadCount = 0 // количество непрочитанных сообщений в чате = 0
	chat.Gist = nil
	return chat, nil
}

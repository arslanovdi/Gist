package core

import (
	"context"
	"fmt"
)

// MarkAsRead отметить сообщения чата как прочитанные. Нумерация страниц начинается с 1.
func (g *Gist) MarkAsRead(ctx context.Context, chatID int64, pageID int) error {

	chat, errD := g.GetChatDetail(ctx, chatID)
	if errD != nil {
		return fmt.Errorf("core.MarkAsRead: %w", errD)
	}

	errM := g.tgClient.MarkAsRead(ctx, chat, chat.Gist[pageID-1].LastMessageID)
	if errM != nil {
		return fmt.Errorf("core.MarkAsRead: %w", errM)
	}

	return nil
}

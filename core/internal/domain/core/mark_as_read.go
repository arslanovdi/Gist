package core

import (
	"context"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// MarkAsRead отметить сообщения чата как прочитанные.
func (g *Gist) MarkAsRead(ctx context.Context, chat *model.Chat, MaxID int) error {

	return g.tgClient.MarkAsRead(ctx, chat, MaxID)
}

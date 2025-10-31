package core

import (
	"context"
	"fmt"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

func (g *Gist) GetChatGist(_ context.Context, chatID int64) (string, error) {
	// TODO implement me
	chat, ok := g.cache[chatID]
	if !ok {
		return "", model.ErrChatNotFoundInCache
	}
	chat.Gist = fmt.Sprintf("implement me, chat_id: %d", chatID)

	return chat.Gist, nil
}

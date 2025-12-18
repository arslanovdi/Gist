package core

import (
	"context"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// GetChatGist возвращает короткий пересказ непрочитанных сообщений чата.
func (g *Gist) GetChatGist(ctx context.Context, chatID int64) (string, error) {
	chat, ok := g.cache[chatID]
	if !ok {
		return "", model.ErrChatNotFoundInCache
	}

	messages, errF := g.tgClient.FetchUnreadMessages(ctx, *chat) // получаем список непрочитанных сообщений чата
	if errF != nil {
		return "", errF
	}

	resp, errG := g.llmClient.GetChatGist(ctx, messages) // Выделяем суть из сообщений
	if errG != nil {
		return "", errG
	}

	chat.Gist = resp

	return chat.Gist, nil
}

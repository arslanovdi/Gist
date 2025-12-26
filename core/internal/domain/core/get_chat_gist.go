package core

import (
	"context"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// GetChatGist возвращает короткий пересказ непрочитанных сообщений чата.
func (g *Gist) GetChatGist(ctx context.Context, chatID int64) ([]model.BatchGist, error) {
	chat, ok := g.cache[chatID]
	if !ok {
		return nil, model.ErrChatNotFoundInCache
	}

	messages, errF := g.tgClient.FetchUnreadMessages(ctx, *chat) // получаем список непрочитанных сообщений чата
	if errF != nil {
		return nil, errF
	}

	resp, errG := g.llmClient.GetChatGist(ctx, messages) // Выделяем суть из сообщений
	if errG != nil {
		return nil, errG
	}

	chat.Gist = resp

	return chat.Gist, nil
}

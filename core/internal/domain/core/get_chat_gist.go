package core

import (
	"context"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// GetChatGist возвращает короткий пересказ непрочитанных сообщений чата. Callback - оповещение пользователя о ходе выполнения.
func (g *Gist) GetChatGist(ctx context.Context, chatID int64, callback func(string, int, bool)) ([]model.BatchGist, error) {
	chat, ok := g.cache[chatID]
	if !ok {
		return nil, model.ErrChatNotFoundInCache
	}

	messages, errF := g.tgClient.FetchUnreadMessages(ctx, chat, callback) // получаем список непрочитанных сообщений чата
	if errF != nil {
		return nil, errF
	}

	resp, errG := g.llmClient.GetChatGist(ctx, messages, callback) // Выделяем суть из сообщений
	if errG != nil {
		return nil, errG
	}

	chat.Gist = resp

	return chat.Gist, nil
}

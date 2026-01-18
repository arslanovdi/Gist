package core

import (
	"context"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// GetChatGist возвращает короткий пересказ непрочитанных сообщений чата. Callback - оповещение пользователя о ходе выполнения.
func (g *Gist) GetChatGist(ctx context.Context, chatID int64, callback func(string, int, bool)) ([]model.BatchGist, error) {

	log := slog.With("func", "core.GetChatGist")

	chat, ok := g.cache[chatID]
	if !ok {
		return nil, model.ErrChatNotFoundInCache
	}

	if chat.Messages == nil {
		messages, skipped, errF := g.tgClient.FetchUnreadMessages(ctx, chat, callback) // получаем список непрочитанных сообщений чата
		if errF != nil {
			return nil, errF
		}
		chat.Messages = messages
		chat.Skipped = skipped
	} else {
		log.Debug("messages already loaded", slog.Int("count", len(chat.Messages)), slog.Int("skipped", chat.Skipped))
	}

	resp, errG := g.llmClient.GenerateChatGist(ctx, chat.Messages, callback) // Выделяем суть из сообщений
	if errG != nil {
		return nil, errG
	}

	chat.Gist = resp

	return chat.Gist, nil
}

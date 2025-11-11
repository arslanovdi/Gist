package core

import (
	"context"
	"fmt"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

func (g *Gist) GetChatGist(ctx context.Context, chatID int64) (string, error) {
	// TODO implement me
	chat, ok := g.cache[chatID]
	if !ok {
		return "", model.ErrChatNotFoundInCache
	}

	messages, errF := g.tgClient.FetchUnreadMessages(ctx, *chat)
	if errF != nil {
		return "", errF
	}

	fmt.Println("new", messages[0]) // TODO implement send to LLM
	fmt.Println("old", messages[len(messages)-1])

	chat.Gist = fmt.Sprintf("implement me, chat_id: %d", chatID)

	return chat.Gist, nil
}

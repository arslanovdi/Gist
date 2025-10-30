package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

func (g *Gist) GetFavoriteChats(ctx context.Context) ([]model.Chat, error) {
	log := slog.With("func", "core.GetFavoriteChats")
	log.Debug("get favorite chats")

	chats, errA := g.GetAllChats(ctx)
	if errA != nil {
		return nil, fmt.Errorf("GetFavoriteChats: %w", errA)
	}

	// Отбрасываем чаты, которых нет в списке Favorites
	if len(chats) == 0 {
		return nil, nil
	}

	favoriteChats := make([]model.Chat, 0)
	for i := range chats {
		if chats[i].IsFavorite {
			favoriteChats = append(favoriteChats, chats[i])
		}
	}

	log.Debug("Successfully get favorite chats", slog.Any("favoriteChats", favoriteChats))

	return favoriteChats, nil

}

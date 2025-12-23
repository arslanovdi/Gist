package core

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// GetAllChats возвращает список всех чатов пользователя.
func (g *Gist) GetAllChats(ctx context.Context) ([]model.Chat, error) {
	log := slog.With("func", "core.GetAllChats")
	log.Debug("Get all chats")

	if time.Since(g.lastUpdate) < g.ttl { // Ходим в кеш, пока не вышел TTL
		return g.chats, nil
	}

	ctxClient, cancelClient := context.WithTimeout(ctx, g.requestTimeout) // Контекст ограничивающий время выполнения запроса (включая закрытие горутин аутентификации в боте и клиенте по тайм-ауту)
	defer cancelClient()

	chats, errG := g.tgClient.GetAllChats(ctxClient)
	if errG != nil {
		return nil, errG // Прочие ошибки
	}

	// отсортировать по убыванию UnreadCount
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].UnreadCount > chats[j].UnreadCount
	})

	// Сохраняем полученные чаты в инмемори
	g.lastUpdate = time.Now()
	g.chats = chats
	g.cache = make(map[int64]*model.Chat)
	for i := range chats {
		g.cache[chats[i].ID] = &chats[i]
	}

	log.Debug("Successfully get all chats", slog.Any("chats", chats))

	return chats, nil
}

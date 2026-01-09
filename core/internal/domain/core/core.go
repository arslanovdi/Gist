// Package core слой бизнес-логики
package core

import (
	"context"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
)

// TelegramClient контракт для работы с телеграмм клиентом
type TelegramClient interface {
	GetAllChats(ctx context.Context) ([]model.Chat, error)
	FetchUnreadMessages(ctx context.Context, chat *model.Chat, callback func(message string, count int, llm bool)) ([]model.Message, int, error)
	MarkAsRead(ctx context.Context, chat *model.Chat, lastMessageID int) error
}

// LLMClient контракт для работы с LLM
type LLMClient interface {
	GenerateChatGist(ctx context.Context, messages []model.Message, callback func(message string, progress int, llm bool)) ([]model.BatchGist, error)
	GenerateAudioGist(ctx context.Context, chat *model.Chat, batchID int) error // Генерирует аудиопересказы по каждому из батчей
}

// Gist представляет ядро бизнес-логики приложения.
type Gist struct {
	tgClient  TelegramClient
	llmClient LLMClient

	cache      map[int64]*model.Chat // Для быстрого доступа TODO вынести кэш в отдельный слой?
	chats      []model.Chat
	lastUpdate time.Time
	ttl        time.Duration

	UnreadThreshold int
	cfg             *config.Config

	requestTimeout time.Duration
}

// ChangeFavorites добавление чата в избранное
func (g *Gist) ChangeFavorites(_ context.Context, chatID int64) error {
	// TODO implement me, save to DB
	chat, ok := g.cache[chatID]
	if !ok {
		return model.ErrChatNotFoundInCache
	}

	chat.IsFavorite = !chat.IsFavorite
	return nil
}

// GetChatDetail Получение информации о чате из кэша
func (g *Gist) GetChatDetail(_ context.Context, chatID int64) (*model.Chat, error) {
	chat, ok := g.cache[chatID]
	if !ok {
		return nil, model.ErrChatNotFoundInCache
	}
	return chat, nil
}

// NewGist конструктор
func NewGist(tgClient TelegramClient, llmClient LLMClient, cfg *config.Config) *Gist {
	return &Gist{
		tgClient:        tgClient,
		llmClient:       llmClient,
		requestTimeout:  cfg.Client.RequestTimeout,
		UnreadThreshold: cfg.Settings.ChatUnreadThreshold,
		ttl:             cfg.Project.TTL,
		cfg:             cfg,
	}
}

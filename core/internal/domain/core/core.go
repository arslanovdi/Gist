package core

import (
	"context"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
)

const ttl = time.Minute

type TelegramClient interface {
	GetAllChats(ctx context.Context) ([]model.Chat, error)
	FetchUnreadMessages(ctx context.Context, chat model.Chat) ([]model.Message, error)
}

type LLMClient interface {
	GetChatGist(ctx context.Context, messages []model.Message) (string, error)
}

// Gist представляет основной сервис приложения, который инкапсулирует бизнес-логику.
type Gist struct {
	tgClient  TelegramClient
	llmClient LLMClient

	cache      map[int64]*model.Chat // Для быстрого доступа TODO вынести кэш в отдельный слой?
	chats      []model.Chat
	lastUpdate time.Time
	ttl        time.Duration

	UnreadThreshold int

	requestTimeout time.Duration
}

func (g *Gist) ChangeFavorites(_ context.Context, chatID int64) error {
	// TODO implement me, save to DB
	chat, ok := g.cache[chatID]
	if !ok {
		return model.ErrChatNotFoundInCache
	}

	chat.IsFavorite = !chat.IsFavorite
	return nil
}

func (g *Gist) GetChatDetail(_ context.Context, chatID int64) (*model.Chat, error) {
	chat, ok := g.cache[chatID]
	if !ok {
		return nil, model.ErrChatNotFoundInCache
	}
	return chat, nil
}

func NewGist(tgClient TelegramClient, llmClient LLMClient, cfg *config.Config) *Gist {
	return &Gist{
		tgClient:        tgClient,
		llmClient:       llmClient,
		requestTimeout:  cfg.Client.RequestTimeout,
		UnreadThreshold: cfg.Settings.ChatUnreadThreshold,
		ttl:             ttl,
	}
}

package tgclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

// GetAllChats возвращает список всех чатов и диалогов пользователя из Telegram.
//
//nolint:gocyclo //cyclo-15
func (s *Session) GetAllChats(ctx context.Context) ([]model.Chat, error) {
	log := slog.With("func", "tgclient.GetAllChats", slog.Int64("user_id", s.userID))
	log.Debug("Get all chats")

	if !s.ready.Load() {
		return nil, model.ErrNotReady
	}

	chats := make([]model.Chat, 0)

	raw := tg.NewClient(s.client)
	builder := query.GetDialogs(
		raw,
	) // Используем хелпер, для получения списка диалогов, с учетом пагинации
	builder.BatchSize(batchLimit)

	elems, err := builder.Collect(ctx) // Получаем все элементы
	if err != nil {
		return nil, fmt.Errorf("collect dialogs failed: %w", err)
	}

	for _, elem := range elems {

		chat := model.Chat{}

		switch d := elem.Dialog.(type) { // Получаем количество непрочитанных сообщений
		case *tg.Dialog:
			chat.UnreadCount = d.UnreadCount
			chat.LastReadMessageID = d.ReadInboxMaxID
		case *tg.DialogFolder:
			log.Info("tg.DialogFolder")
		case nil:
			log.Error("nil dialog")
		default:
			log.Error("Unknown peer type")
		}

		switch peer := elem.Peer.(type) {
		case *tg.InputPeerChat:
			chat.ID = peer.ChatID
			chat.Title = elem.Entities.Chats()[chat.ID].Title
			chat.Peer = peer
		case *tg.InputPeerUser:
			chat.ID = peer.UserID
			chat.Title = elem.Entities.Users()[chat.ID].Username
			chat.Peer = peer
		case *tg.InputPeerChannel:
			chat.ID = peer.ChannelID
			chat.Title = elem.Entities.Channels()[chat.ID].Title
			chat.Peer = peer
		case *tg.InputPeerEmpty:
			log.Info("tg.InputPeerEmpty")
		case *tg.InputPeerSelf:
			log.Info("tg.InputPeerSelf")
		case *tg.InputPeerUserFromMessage:
			log.Info("tg.InputPeerUserFromMessage")
		case *tg.InputPeerChannelFromMessage:
			log.Info("tg.InputPeerChannelFromMessage")
		case nil:
			log.Error("nil peer")
		default:
			log.Error("Unknown peer type")
		}

		chats = append(chats, chat)
	}

	log.Debug("Get all chats done")

	return chats, nil
}

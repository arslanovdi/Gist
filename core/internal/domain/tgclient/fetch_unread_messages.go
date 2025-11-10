package tgclient

import (
	"context"
	"log/slog"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

func (s *Session) FetchUnreadMessages(ctx context.Context, chat model.Chat) ([]model.Message, error) {
	log := slog.With("func", "tgclient.FetchUnreadMessages", slog.Any("chatID", chat))
	log.Debug("Get unread messages from chat")

	if !s.ready.Load() {
		return nil, model.ErrNotReady
	}

	msgs := make([]model.Message, 0)

	raw := tg.NewClient(s.client)
	builder := query.Messages(raw)

	historyBuilder := builder.GetHistory(chat.Peer)
	historyBuilder.BatchSize(batchLimit)

	iter := historyBuilder.Iter()

	for iter.Next(ctx) {
		// пропускаем сервисные сообщения
		if _, ok := iter.Value().Msg.(*tg.MessageService); ok {
			continue
		}

		tgMsg, ok := iter.Value().Msg.(*tg.Message)
		if !ok {
			log.Error("Got message with unexpected type", slog.Any("msg_type", iter.Value().Msg))
			continue
		}

		if tgMsg.ID <= chat.LastReadMessageID {
			break
		}

		message := model.Message{
			ID:          tgMsg.ID,
			Text:        tgMsg.Message,
			Timestamp:   time.Unix(int64(tgMsg.Date), 0),
			IsEdited:    tgMsg.EditDate > 0,
			MessageType: "text", // По умолчанию
		}

		msgs = append(msgs, message)

	}

	log.Debug("Get unread messages done", slog.Int("count", len(msgs)))

	return msgs, nil
}

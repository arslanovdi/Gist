package tgclient

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

// FetchUnreadMessages выгружает непрочитанные сообщения из телеграмм чата
// callback - оповещение пользователя о ходе выполнения.
//
//nolint:gocognit,gocyclo // cognit-21, cyclo-15
func (s *Session) FetchUnreadMessages(ctx context.Context, chat *model.Chat, callback func(message string, count int, llm bool)) ([]model.Message, error) {
	log := slog.With(slog.String("func", "tgclient.FetchUnreadMessages"), slog.Any("chatID", chat))
	log.Debug("Get unread messages from chat")

	if !s.ready.Load() {
		return nil, model.ErrNotReady
	}

	callback("📥 Загружаем сообщения из Telegram...", 0, false) // Оповещение пользователю в телеграм бот
	ticker := time.Now()

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
			log.Error("Got message with unexpected type", slog.Any("type", iter.Value().Msg))

			continue
		}

		// Читаем только новые сообщения.
		if tgMsg.ID <= chat.LastReadMessageID {
			break
		}

		// Пропускаем пустые сообщения (без текста).
		if tgMsg.Message == "" {
			continue
		}

		message := model.Message{
			ID:           tgMsg.ID,
			Text:         tgMsg.Message,
			Timestamp:    time.Unix(int64(tgMsg.Date), 0),
			IsEdited:     tgMsg.EditDate > 0,
			SenderID:     0,     // Заполняется дальше
			ReplyToMsgID: 0,     // Заполняется дальше
			IsForwarded:  false, // Заполняется дальше
		}

		if _, ok := tgMsg.GetFwdFrom(); ok {
			message.IsForwarded = true
		}

		// Заполняем SenderID
		if peerClass, ok := tgMsg.GetFromID(); ok {
			switch fromID := peerClass.(type) {
			case *tg.PeerUser:
				message.SenderID = fromID.UserID
			case *tg.PeerChat:
				message.SenderID = fromID.ChatID
			case *tg.PeerChannel:
				message.SenderID = fromID.ChannelID
			case nil:
				log.Error("FromID type is nil")
			default:
				log.Error("FromID type is unknown")
			}
		}

		// Заполняем ReplyToMsgID
		if messageReply, ok := tgMsg.GetReplyTo(); ok {
			if replyHeader, ok := messageReply.(*tg.MessageReplyHeader); ok {
				message.ReplyToMsgID = replyHeader.ReplyToMsgID
			}
		}

		msgs = append(msgs, message)

		if time.Since(ticker) > time.Second {
			callback("📥 Загружаем сообщения из Telegram...", len(msgs), false) // Оповещение пользователю в телеграм бот
			ticker = time.Now()
		}
	}

	// Проверяем финальную ошибку итератора
	if iter.Err() != nil {
		return nil, fmt.Errorf("tgclient.FetchUnreadMessages failed to iterate messages: %w", iter.Err())
	}

	log.Debug("Get unread messages done", slog.Int("count", len(msgs)))

	// Возвращаем сообщения в правильной хронологии, так как вычитывали их с конца.
	slices.SortFunc(msgs, func(a, b model.Message) int {
		return a.ID - b.ID
	})

	return msgs, nil
}

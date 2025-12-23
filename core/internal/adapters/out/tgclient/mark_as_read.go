package tgclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/gotd/td/tg"
)

// MarkAsRead отмечает чат как прочитанный, MaxID - id сообщения, до которого все считаются прочитанными
func (s *Session) MarkAsRead(ctx context.Context, chat *model.Chat, MaxID int) error {

	log := slog.With("func", "tgclient.MarkAsRead")
	log.Debug("Mark As Read", slog.Int64("chat_id", chat.ID), slog.Int("max_id", MaxID))

	switch peer := chat.Peer.(type) {
	case *tg.InputPeerChat, *tg.InputPeerUser:
		req := &tg.MessagesReadHistoryRequest{
			Peer:  chat.Peer,
			MaxID: MaxID,
		}

		_, err := s.client.API().MessagesReadHistory(ctx, req) // TODO проверить
		if err != nil {
			return err
		}
		log.Debug("Mark As Read Done InputPeerChat / InputPeerUser")

	case *tg.InputPeerChannel:
		req := &tg.ChannelsReadHistoryRequest{
			Channel: &tg.InputChannel{ChannelID: peer.ChannelID, AccessHash: peer.AccessHash},
			MaxID:   MaxID,
		}

		_, err := s.client.API().ChannelsReadHistory(ctx, req)
		if err != nil {
			return err
		}
		log.Debug("Mark As Read Done InputPeerChannel")

	case *tg.InputPeerEmpty:
		return fmt.Errorf("tg.InputPeerEmpty")
	case *tg.InputPeerSelf:
		return fmt.Errorf("tg.InputPeerSelf")
	case *tg.InputPeerUserFromMessage:
		return fmt.Errorf("tg.InputPeerUserFromMessage")
	case *tg.InputPeerChannelFromMessage:
		return fmt.Errorf("tg.InputPeerChannelFromMessage")
	case nil:
		return fmt.Errorf("chat.Peer = nil")
	default:
		return fmt.Errorf("unknown peer type")
	}

	return nil
}

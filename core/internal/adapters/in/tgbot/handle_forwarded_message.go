package tgbot

import (
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// HandleForwardedMessage обрабатывает пересланные сообщения для добавления чатов в избранное.
// Поддерживает пересылку из каналов, групповых чатов и личных сообщений.
func (b *Bot) HandleForwardedMessage(ctx *th.Context, message telego.Message) error {
	log := slog.With("func", "tgbot.HandleForwardedMessage")
	log.Debug("")

	if message.ForwardOrigin == nil {
		return nil
	}

	forwardID := int64(0)
	switch f := message.ForwardOrigin.(type) {
	case *telego.MessageOriginChannel:
		forwardID = f.Chat.ID
	case *telego.MessageOriginChat:
		forwardID = f.SenderChat.ID
	case *telego.MessageOriginUser:
		// пересланное сообщение пользователя не хранит ID чата, в котором оно было написано!
	case *telego.MessageOriginHiddenUser:
	}

	errF := b.coreService.ChangeFavorites(ctx, forwardID)
	if errF != nil {
		return fmt.Errorf("add chat to favorites error: %w", errF)
	}

	log.Debug("chat added to favorites", slog.Int64("chat_id", forwardID))
	return nil
}

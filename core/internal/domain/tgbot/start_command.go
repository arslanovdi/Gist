package tgbot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// StartCommand обрабатывает команду /start от пользователя.
func (b *Bot) StartCommand(ctx *th.Context, update telego.Update) error {

	log := slog.With("func", "tgbot.StartCommand")
	log.Debug("/start command")

	chats, errG := b.coreService.GetChatsWithUnreadMessages(ctx) // Получаем список чатов с непрочитанными сообщениями
	if errG != nil {
		return fmt.Errorf("internal server error: %w", errG)
	}

	err := b.sendChatsWithUnreadMessages(ctx, update, chats) // Выводим результат в чат с пользователем.
	if err != nil {
		log.Error("Failed to send chat list", slog.Any("error", err))
		return fmt.Errorf("failed to send chat list: %w", err)
	}

	log.Debug("/start command finished")

	return nil
}

// sendChatsWithUnreadMessages отправляет пользователю список чатов с количеством непрочитанных сообщений.
func (b *Bot) sendChatsWithUnreadMessages(ctx context.Context, update telego.Update, chats []model.Chat) error {
	const maxLen = 4096
	var message strings.Builder

	sendChunk := func() error {
		if message.Len() == 0 {
			return nil
		}
		_, err := b.bot.SendMessage(ctx, tu.Message(
			tu.ID(update.Message.Chat.ID),
			message.String(),
		))
		message.Reset()
		return err
	}

	for _, chat := range chats {
		line := fmt.Sprintf("%s %d\n", chat.Title, chat.UnreadCount)
		if message.Len()+len(line) > maxLen {
			if err := sendChunk(); err != nil {
				return err
			}
		}
		message.WriteString(line)
	}

	return sendChunk()
}

package tgbot

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// StartCommand обрабатывает команду /start от пользователя.
func (b *Bot) StartCommand(ctx *th.Context, update telego.Update) error {

	b.LastMessageID = update.Message.MessageID // Сохранили номер сообщения

	log := slog.With("func", "tgbot.StartCommand")
	log.Debug("/start command")

	b.showMainMenu(ctx)

	log.Debug("/start command finished")

	return nil
}

/*// sendChatsWithUnreadMessages отправляет пользователю список чатов с количеством непрочитанных сообщений.
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
}*/

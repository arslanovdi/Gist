package tgbot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (b *Bot) StartCommand(ctx *th.Context, update telego.Update) error {

	log := slog.With("func", "tgbot.StartCommand")
	log.Debug("/start command")

	chats, errG := b.coreService.GetAllChats(ctx)
	if errG != nil {
		return errG
	}

	fmt.Println(chats)

	err := b.sendChatsList(ctx, update, chats)
	if err != nil {
		log.Error("Failed to send chat list")
	}

	_, errS := b.bot.SendMessage(ctx, tu.Messagef(
		tu.ID(update.Message.Chat.ID),
		"Hello %s!", update.Message.From.FirstName,
	))
	return errS
}

func (b *Bot) sendChatsList(ctx context.Context, update telego.Update, chats []string) error {
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
		line := chat + "\n" // добавляем перенос
		if message.Len()+len(line) > maxLen {
			if err := sendChunk(); err != nil {
				return err
			}
		}
		message.WriteString(line)
	}

	return sendChunk()
}

package tgbot

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// AnyCommand обработка всех команд, кроме /start
func (b *Bot) AnyCommand(ctx *th.Context, update telego.Update) error {
	_, err := b.bot.SendMessage(ctx, tu.Messagef(
		tu.ID(update.Message.Chat.ID),
		"Unknown command, use /start",
	))
	return err
}

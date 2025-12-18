package tgbot

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// StartCommand обрабатывает команду /start от пользователя.
func (b *Bot) StartCommand(ctx *th.Context, _ telego.Update) error {

	log := slog.With("func", "tgbot.StartCommand")
	log.Debug("/start command")

	err := b.router.ShowMainMenu(ctx)
	if err != nil {
		return err
	}

	log.Debug("/start command finished")

	return nil
}

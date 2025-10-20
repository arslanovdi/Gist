package tgbot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (b *Bot) RegisterHandlers(_ context.Context, serverErr chan error) {
	log := slog.With("func", "tgbot.RegisterHandlers")
	log.Info("Register handlers start")

	var errH error
	b.bh, errH = th.NewBotHandler(b.bot, b.updates)
	if errH != nil {
		serverErr <- fmt.Errorf("[tgbot.RegisterHandlers] error creating handler: %w", errH)
		return
	}

	// Если команда соответствует нескольким обработчикам сработает первый из зарегистрированных
	// сначала нужно определять частные, затем общие обработчики.

	b.bh.Use(th.PanicRecovery()) // Обертка PanicRecovery, вызывается первым.
	b.bh.Handle(b.StartCommand, th.CommandEqual("start"))
	b.bh.Handle(b.AnyCommand, th.AnyCommand())

	b.bh.HandleMessage(b.AuthMessageHandler, th.AnyMessage())
	b.bh.HandleMessage(b.AnyMessage, th.AnyMessage())

	go func() {
		errS := b.bh.Start()
		if errS != nil {
			serverErr <- fmt.Errorf("[tgbot.RegisterHandlers] error starting handler: %w", errS)
		}
		log.Info("Handlers stopped")
	}()

	log.Info("Register handlers successfully")
}

func (b *Bot) AnyCommand(ctx *th.Context, update telego.Update) error {
	_, err := b.bot.SendMessage(ctx, tu.Messagef(
		tu.ID(update.Message.Chat.ID),
		"Unknown command, use /start",
	))
	return err
}

func (b *Bot) AnyMessage(_ *th.Context, message telego.Message) error {
	fmt.Println("Message:", message.Text)
	return nil
}

package tgbot

import (
	"context"
	"fmt"
	"log/slog"

	th "github.com/mymmrac/telego/telegohandler"
)

// RegisterHandlers регистрирует обработчики команд и сообщений для бота.
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

	b.bh.Use(th.PanicRecovery())           // Обертка PanicRecovery, вызывается первым.
	b.bh.Use(b.authorizedOnlyMiddleware()) // Обертка отсеивает запросы всех пользователей, кроме ID указанного в env.
	b.bh.Handle(b.StartCommand, th.CommandEqual("start"))
	b.bh.Handle(b.AnyCommand, th.AnyCommand())

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

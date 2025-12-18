package tgbot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/adapters/in/tgbot/router"
	th "github.com/mymmrac/telego/telegohandler"
)

// RegisterHandlers регистрирует обработчики команд и сообщений для бота.
func (b *Bot) RegisterHandlers(_ context.Context, serverErr chan error) {
	log := slog.With("func", "tgbot.RegisterHandlers")
	log.Info("Register handlers start")

	base := &router.BaseHandler{
		Bot:         b.bot,
		CoreService: b.coreService,
		UserID:      b.allowedUserID,
	}

	// menu
	b.router.RegisterHandler(router.NewMainMenuHandler(base))
	b.router.RegisterHandler(router.NewUnreadMenuHandler(base))
	b.router.RegisterHandler(router.NewFavoritesMenuHandler(base))
	b.router.RegisterHandler(router.NewChatMenuHandler(base))
	b.router.RegisterHandler(router.NewSettingsMenuHandler(base))
	// actions
	b.router.RegisterHandler(router.NewAddToFavoritesHandler(base))
	b.router.RegisterHandler(router.NewTTSHandler(base))
	b.router.RegisterHandler(router.NewMarkAsReadHandler(base))

	var errH error
	b.bh, errH = th.NewBotHandler(b.bot, b.updates)
	if errH != nil {
		serverErr <- fmt.Errorf("[tgbot.RegisterHandlers] error creating handler: %w", errH)
		return
	}

	// Если команда соответствует нескольким обработчикам сработает первый из зарегистрированных
	// сначала нужно определять частные, затем общие обработчики.

	// middlewares
	b.bh.Use(th.PanicRecovery())           // Обертка PanicRecovery, вызывается первым.
	b.bh.Use(b.authorizedOnlyMiddleware()) // Обертка отсеивает запросы всех пользователей, кроме ID указанного в env.

	// commands
	b.bh.Handle(b.StartCommand, th.CommandEqual("start"))
	b.bh.Handle(b.AnyCommand, th.AnyCommand())

	// messages
	b.bh.HandleMessage(b.HandleForwardedMessage, th.AnyMessage())

	// callback-запросы (инлайн-кнопки), вызываем обработчик роутера
	b.bh.HandleCallbackQuery(b.router.Handle, th.AnyCallbackQuery())

	go func() {
		errS := b.bh.Start()
		if errS != nil {
			serverErr <- fmt.Errorf("[tgbot.RegisterHandlers] error starting handler: %w", errS)
		}
		log.Info("Handlers stopped")
	}()

	log.Info("Register handlers successfully")
}

package tgbot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/mymmrac/telego"
	"github.com/valyala/fasthttp"
	"golang.ngrok.com/ngrok/v2"
)

type UserCommands interface {
	GetAllChats() ([]string, error)
	GetGistFromChat(chatID int) (string, error)
}

type Bot struct {
	cfg     *config.Config
	bot     *telego.Bot
	srv     *fasthttp.Server
	agent   ngrok.Agent
	tun     ngrok.EndpointListener
	updates <-chan telego.Update
	wg      *sync.WaitGroup
	stop    chan struct{}

	commands UserCommands
}

func New(cfg *config.Config, commands UserCommands) (*Bot, error) {
	// log := slog.With("func", "bot.New")

	// создаем бота
	bot, errB := telego.NewBot(cfg.Bot.Token,
		telego.WithDefaultDebugLogger()) // TODO change DefaultDebugLogger Note: Please keep in mind that default logger may expose sensitive information, use in development only
	if errB != nil {
		return nil, fmt.Errorf("[bot.New] bot initialization failed: %w", errB)
	}

	// создаем агента
	// TODO после получения хостинга / белого IP + сертификата: отказ от ngrok?
	agent, errA := ngrok.NewAgent(ngrok.WithAuthtoken(cfg.Bot.NgrokAuthToken))
	if errA != nil {
		return nil, fmt.Errorf("[bot.New] ngrok agent initialization failed: %w", errA)
	}

	// Создаем сервер, для обработки запросов вебхука Telegram
	srv := &fasthttp.Server{}

	return &Bot{
		bot:      bot,
		srv:      srv,
		agent:    agent,
		commands: commands,
		cfg:      cfg,
		wg:       &sync.WaitGroup{},
		stop:     make(chan struct{}),
	}, nil
}

// Run запускает Telegram-бота, начиная обработку вебхуков и обновлений.
func (b *Bot) Run(ctx context.Context, serverErr chan error) {
	log := slog.With("func", "bot.Run")

	// создаем туннель
	var errT error

	b.tun, errT = b.agent.Listen(ctx,
		ngrok.WithURL(b.cfg.Bot.NgrokDomain))
	if errT != nil {
		serverErr <- fmt.Errorf("[bot.Run] ngrok listen failed: %w", errT)
		return
	}

	// Start server for receiving requests from the Telegram using the Ngrok tunnel
	// TODO Запускаем fasthttp сервер на ngrok-туннеле
	b.wg.Go(func() {
		errS := b.srv.Serve(b.tun)
		if errS != nil { // TODO будет ли выводится ошибка при корректной остановке сервера, может переделать на net/http, там можно обработать корректное завершение.
			serverErr <- fmt.Errorf("[bot.Run] bot webhook serve failed: %w", errS)
		}
	})

	// Регистрируем вебхук, получаем канал обновлений, используя Ngrok
	var errB error
	b.updates, errB = b.bot.UpdatesViaWebhook(ctx,
		// Use FastHTTP webhook server
		telego.WebhookFastHTTP(b.srv, "/bot", b.bot.SecretToken()),
		// Calls SetWebhook before starting webhook and provide dynamic Ngrok tunnel URL
		telego.WithWebhookSet(ctx, &telego.SetWebhookParams{
			URL:         b.tun.URL().String() + "/bot",
			SecretToken: b.bot.SecretToken(),
		}),
	)
	if errB != nil {
		serverErr <- fmt.Errorf("[bot.Run] bot updates via webhook failed: %w", errB)
		return
	}

	// Loop through all updates when they came
	b.wg.Go(func() {
		for {
			select {
			case update, ok := <-b.updates:
				if !ok {
					return
				}
				messageProcessing(update)
			case <-b.stop: // Так как закрытие канала update происходит через метод b.bot.Close, а его можно вызывать только через 10 минут после запуска. Реализован альтернативный способ закрытия канала через сигнал.
				return
			}
		}
	})

	log.Info("bot started")
}

func (b *Bot) Close(ctx context.Context) {
	log := slog.With("func", "bot.Close")
	log.Info("Stopping...")

	errT := b.tun.CloseWithContext(ctx)
	if errT != nil {
		log.Error("Error shutting down tunnel", slog.Any("error", errT))
	}

	errS := b.srv.ShutdownWithContext(ctx)
	if errS != nil {
		log.Error("Error shutting down server", slog.Any("error", errS))
	}

	errD := b.bot.DeleteWebhook(ctx, &telego.DeleteWebhookParams{}) // TODO удалять вебхук в принципе необязательно. Если не удалить longPolling работать не будет.
	if errD != nil {
		log.Error("Delete webhook error", slog.Any("error", errD))
	}

	// Очищаем очередь обновлений
	fmt.Println("len", len(b.updates))
loop:
	for len(b.updates) > 0 {
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(time.Microsecond * 100):
			fmt.Println("tik")
			// Continue
		}
	}

	// TODO Метод Close нельзя вызывать раньше чем через 10 минут после запуска.! Подумать для чего он вообще вызывается.
	/*errB := b.bot.Close(ctx)
	if errB != nil {
		log.Error("Error shutting down bot", slog.Any("error", errB))
	}*/

	close(b.stop) // Отправляем сигнал на завершение обработки updates.

	b.wg.Wait()
	log.Info("Bot Stopped")
}

func messageProcessing(update telego.Update) {
	fmt.Printf("Update: %+v\n", update)
}

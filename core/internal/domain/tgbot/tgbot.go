package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"golang.ngrok.com/ngrok/v2"
)

// CoreService определяет интерфейс для взаимодействия с бизнес-логикой.
type CoreService interface {
	GetAllChats(ctx context.Context) ([]model.Chat, error)
	GetChatsWithUnreadMessages(ctx context.Context) ([]model.Chat, error)
	GetFavoriteChats(ctx context.Context) ([]model.Chat, error)
	GetChatGist(ctx context.Context, chatID int64) (string, error)
	GetChatDetail(ctx context.Context, chatID int64) (*model.Chat, error)
	ChangeFavorites(ctx context.Context, chatID int64) error
}

type Bot struct {
	cfg           *config.Config
	allowedUserID int64

	LastMessageID int // id редактируемого сообщения. В боте всегда одно сообщение, которое мы редактируем.

	// параметры ngrok туннеля
	srv   *http.Server
	agent ngrok.Agent
	tun   ngrok.EndpointListener

	bot     *telego.Bot // параметры телеграм бота
	bh      *th.BotHandler
	updates <-chan telego.Update

	wg *sync.WaitGroup // Контроль запущенных горутин (веб-сервер)

	coreService CoreService // Слой бизнес-логики
}

func New(cfg *config.Config, coreService CoreService) (*Bot, error) {
	log := slog.With("func", "bot.New")
	log.Info("Initializing bot")

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

	srv := &http.Server{}

	return &Bot{
		bot:           bot,
		srv:           srv,
		agent:         agent,
		cfg:           cfg,
		coreService:   coreService,
		wg:            &sync.WaitGroup{},
		allowedUserID: cfg.Client.UserID,
	}, nil
}

// Run запускает Telegram-бота, начиная обработку вебхуков и обновлений.
func (b *Bot) Run(ctx context.Context, serverErr chan error) {
	log := slog.With("func", "bot.Run")

	if serverErr == nil {
		log.Error("error channel is nil")
		return
	}

	// создаем туннель
	var errT error

	b.tun, errT = b.agent.Listen(ctx,
		ngrok.WithURL(b.cfg.Bot.NgrokDomain))
	if errT != nil {
		serverErr <- fmt.Errorf("[bot.Run] ngrok listen failed: %w", errT)
		return
	}

	// Запускаем http сервер на ngrok-туннеле для получения запросов от Telegram
	b.wg.Go(func() {
		errS := b.srv.Serve(b.tun)
		if errS != nil {
			if !errors.Is(errS, http.ErrServerClosed) {
				serverErr <- fmt.Errorf("[bot.Run] bot webhook serve failed: %w", errS)
			}
		}
	})

	// Регистрируем вебхук, получаем канал обновлений, используя Ngrok
	var errB error
	b.updates, errB = b.bot.UpdatesViaWebhook(ctx,
		// Use net/http webhook server
		telego.WebhookHTTPServer(b.srv, "/bot", b.bot.SecretToken()),
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

	b.RegisterHandlers(ctx, serverErr) // Регистрируем все обработчики

	log.Info("bot started")
}

func (b *Bot) Close(ctx context.Context) {
	log := slog.With("func", "bot.Close")
	log.Info("Stopping...")

	errD := b.bot.DeleteWebhook(ctx, &telego.DeleteWebhookParams{})
	if errD != nil {
		log.Error("Delete webhook error", slog.Any("error", errD))
	}

	errH := b.bh.Stop()
	if errH != nil {
		log.Error("Error stopping handling of updates", slog.Any("error", errH))
	}

	errS := b.srv.Shutdown(ctx)
	if errS != nil {
		log.Error("Error shutting down server", slog.Any("error", errS))
	}

	errT := b.tun.CloseWithContext(ctx)
	if errT != nil {
		log.Error("Error shutting down tunnel", slog.Any("error", errT))
	}

	// TODO Метод Close нельзя вызывать раньше чем через 10 минут после запуска.! Подумать для чего он вообще вызывается.
	// API response close: Ok: false, Err: [429 "Too Many Requests: retry after 591", migrate to chat ID: 0, retry after: 591]
	/*errB := b.bot.Close(ctx)
	if errB != nil {
		log.Error("Error shutting down bot", slog.Any("error", errB))
	}*/

	b.wg.Wait()
	log.Info("Bot Stopped")
}

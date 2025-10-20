package tgbot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"github.com/valyala/fasthttp"
	"golang.ngrok.com/ngrok/v2"
)

type CoreService interface {
	GetAllChats(ctx context.Context, userID int64) ([]string, error) // TODO все чаты в боте не нужны, только для теста. В релизе выводить только чаты, в которых есть непрочитанные сообщения, в порядке убывания кол-ва непрочитанных сообщений
}

type Bot struct {
	cfg *config.Config

	srv   *fasthttp.Server // параметры ngrok туннеля
	agent ngrok.Agent
	tun   ngrok.EndpointListener

	bot     *telego.Bot // параметры телеграм бота
	bh      *th.BotHandler
	updates <-chan telego.Update

	wg *sync.WaitGroup // Контроль запущенных горутин (веб-сервер)

	userStates      map[int64]model.UserState   // Состояние пользователя, пользователи удаляются после отправки данных в канал регистрации TODO добавить метрики на кол-во элементов в мапе
	userCredentials map[int64]*model.Credential // Учетные данные, пользователи удаляются после отправки данных в канал регистрации TODO добавить метрики на кол-во элементов в мапе

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
	srv := &fasthttp.Server{
		//	CloseOnShutdown: true,
	}

	return &Bot{
		bot:             bot,
		srv:             srv,
		agent:           agent,
		cfg:             cfg,
		coreService:     coreService,
		wg:              &sync.WaitGroup{},
		userStates:      make(map[int64]model.UserState),
		userCredentials: make(map[int64]*model.Credential),
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

	b.RegisterHandlers(ctx, serverErr) // Регистрируем все обработчики

	log.Info("bot started")
}

func (b *Bot) Close(ctx context.Context) {
	log := slog.With("func", "bot.Close")
	log.Info("Stopping...")

	errD := b.bot.DeleteWebhook(ctx, &telego.DeleteWebhookParams{}) // TODO удалять вебхук в принципе необязательно? Если не удалить longPolling работать не будет.
	if errD != nil {
		log.Error("Delete webhook error", slog.Any("error", errD))
	}

	errH := b.bh.Stop()
	if errH != nil {
		log.Error("Error stopping handling of updates", slog.Any("error", errH))
	}

	errS := b.srv.ShutdownWithContext(ctx)
	if errS != nil {
		log.Error("Error shutting down server", slog.Any("error", errS))
	}

	errT := b.tun.CloseWithContext(ctx)
	if errT != nil {
		log.Error("Error shutting down tunnel", slog.Any("error", errT))
	}

	// TODO Метод Close нельзя вызывать раньше чем через 10 минут после запуска.! Подумать для чего он вообще вызывается.
	/*errB := b.bot.Close(ctx)
	if errB != nil {
		log.Error("Error shutting down bot", slog.Any("error", errB))
	}*/

	b.wg.Wait()
	log.Info("Bot Stopped")
}

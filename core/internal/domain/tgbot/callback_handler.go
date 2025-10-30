package tgbot

import (
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func (b *Bot) handleCallback(ctx *th.Context, query telego.CallbackQuery) error {
	log := slog.With("func", "tgbot.handleCallback")
	log.Debug("handleCallback")

	// Всё, что вам нужно — уже в query
	if query.Data == "" {
		log.Debug("handleCallback: query.Data is empty")
		return fmt.Errorf("no callback data found")
	}

	// Парсим payload
	payload, err := ParseCallback(query.Data)
	if err != nil {
		_ = b.bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "⚠️ Некорректные данные",
			ShowAlert:       true,
		})
		return fmt.Errorf("parse callback data err: %w", err)
	}

	fmt.Println(payload)

	switch {
	// Переход на главное меню
	case payload.Menu == MenuMain:
		b.showMainMenu(ctx)

	// Вывод списка непрочитанных чатов
	case payload.Menu == MenuUnread:
		chats, errU := b.coreService.GetChatsWithUnreadMessages(ctx)
		if errU != nil {
			log.Error("GetChatsWithUnreadMessages", slog.Any("error", errU))
		}

		b.showUnreadChats(ctx, chats, payload.Page)

	// Вывод списка избранных чатов.
	case payload.Menu == MenuFavorites:
		chats, errF := b.coreService.GetFavoriteChats(ctx)
		if errF != nil {
			log.Error("GetFavoriteChats", slog.Any("error", errF))
		}

		b.showFavoriteChats(ctx, chats, payload.Page)

	// Вывод информации по выбранному чату
	case payload.Menu == MenuChat:
		_, errG := b.coreService.GetChatGist(ctx, payload.ChatID) // Метод сохраняет суть в структуру Detail
		if errG != nil {
			log.Error("GetChatGist", slog.Any("error", errG))
		}

		chatDetail, errD := b.coreService.GetChatDetail(ctx, payload.ChatID)
		if errD != nil {
			chatDetail = &model.Chat{}
			log.Error("GetChatDetail", slog.Any("error", errD))
		}

		b.showChatDetail(ctx, *chatDetail, payload.Src)

	case payload.Menu == MenuSettings:
		// TODO implement me
	case payload.Action == ActionMarkRead:
		// TODO implement me
		/*b.markAsRead(payload.ChatID)
		b.answerCallback(cb.ID, "✅ Прочитано!")
		// Обновите сообщение или вернитесь назад*/
	case payload.Action == ActionTTS:
		// TODO implement me
	case payload.Action == ActionToggleFav:
		errF := b.coreService.ChangeFavorites(ctx, payload.ChatID)
		if errF != nil {
			log.Error("ChangeFavorites", slog.Any("error", errF))
		}

		chatDetail, errD := b.coreService.GetChatDetail(ctx, payload.ChatID)
		if errD != nil {
			chatDetail = &model.Chat{}
			log.Error("GetChatDetail", slog.Any("error", errD))
		}

		b.showChatDetail(ctx, *chatDetail, payload.Src)

	default:
		errA := b.bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "🤔 Неизвестное действие",
			ShowAlert:       true,
		})
		if errA != nil {
			log.Error("default AnswerCallbackQuery", slog.Any("error", errA))
		}
	}

	return nil
}

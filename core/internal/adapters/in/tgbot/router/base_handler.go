package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const (
	chatsPerPage = 8 // Количество чатов выводимых пользователю за раз (пагинация)
)

// BaseHandler содержит общие зависимости и методы для всех обработчиков
type BaseHandler struct {
	Bot         *telego.Bot
	CoreService CoreService

	LastMessageID int   // Id редактируемого сообщения. В боте всегда одно сообщение, которое мы редактируем.
	UserID        int64 // Id пользователя = id чата с ним, используется для вывода сообщений ботом.
}

func (b *BaseHandler) showChatDetail(ctx context.Context, chat *model.Chat, menu Menu, gistPage int) error {
	log := slog.With("func", "router.showChatDetail")

	inlineKeyboard := b.buildChatDetailMenu(chat, menu, gistPage)

	text := "" // Текст сообщения. Краткий пересказ выводится только если он сделан.
	if len(chat.Gist) > 0 {
		text = fmt.Sprintf("📩 %s\n🔍 Краткий пересказ %d/%d сообщений:\n\n %s\n", // 📌 Непрочитано: %d сообщений
			chat.Title,
			chat.Gist[gistPage-1].MessageCount,
			chat.UnreadCount,
			chat.Gist[gistPage-1].Gist,
		)
	} else {
		text = fmt.Sprintf("📩 %s\n\n 📌 Непрочитано: %d сообщений",
			chat.Title,
			chat.UnreadCount,
		)
	}

	if b.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(b.UserID),
			b.LastMessageID,
			text,
		).WithReplyMarkup(inlineKeyboard)

		_, errE := b.Bot.EditMessageText(ctx, message)
		if errE == nil {
			return nil // Успешно отредактировали
		}
		log.Error("edit message with chat detail menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(b.UserID),
		text,
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := b.Bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with chat detail menu error", slog.Any("error", errS))
		return fmt.Errorf("send message with chat detail menu error: %w", errS)
	}

	b.LastMessageID = msg.MessageID // Сохраняем номер сообщения
	return nil
}

// Создание меню для выбранного чата.
// gistPage нумерация с 1.
func (b *BaseHandler) buildChatDetailMenu(chat *model.Chat, menu Menu, gistPage int) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Кнопки Назад, Далее для перелистывания страниц с кратким пересказом.
	// Кнопка Назад, активна когда gistPage > 1.
	// Кнопка Вперед активна когда gistPage < len(chat.Gist) // меньше количества страниц кратких пересказов.
	if len(chat.Gist) > 1 {
		backwardGistCb := mustCallback(CallbackPayload{
			ChatID: chat.ID,
			Menu:   MenuChat, // По этому параметру будет выбран обработчик кнопки.
			Src:    menu,     // Меню, из которого вызвано описание чата. Нужна для корректной отработки кнопки "Назад к чатам"
			Page:   gistPage - 1})

		forwardGistCb := mustCallback(CallbackPayload{
			ChatID: chat.ID,
			Menu:   MenuChat,
			Src:    menu,
			Page:   gistPage + 1})

		switch gistPage {
		case 1: // Есть только кнопка Вперед.
			rows = append(rows, tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("→ Вперед").WithCallbackData(forwardGistCb),
			))
		case len(chat.Gist): // Есть только кнопка Назад
			rows = append(rows, tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("← Назад").WithCallbackData(backwardGistCb),
			))
		default: // Есть обе кнопки
			rows = append(rows, tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("← Назад").WithCallbackData(backwardGistCb),
				tu.InlineKeyboardButton("→ Вперед").WithCallbackData(forwardGistCb),
			))
		}
	}

	// Кнопка Пометить прочитанным
	markReadCb := mustCallback(CallbackPayload{
		Action: ActionMarkRead,
		Src:    menu,
		ChatID: chat.ID,
		Page:   gistPage})
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("✅ Пометить прочитанным").WithCallbackData(markReadCb),
	))

	// Кнопка Сгенерировать пересказ
	getGistCb := mustCallback(CallbackPayload{
		Action: ActionGetGist,
		ChatID: chat.ID,
		Src:    menu,
	})
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("✨ Сгенерировать пересказ").WithCallbackData(getGistCb),
	))

	// TODO Кнопка Озвучить
	if len(chat.Gist) > 0 {
		ttsCb := mustCallback(CallbackPayload{
			Action: ActionTTS,
			ChatID: chat.ID})
		rows = append(rows, tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("🔊 Озвучить").WithCallbackData(ttsCb),
		))
	}

	// Кнопка "в избранное" / "убрать из избранного"
	favLabel := "⭐ В избранное"
	add := true
	if chat.IsFavorite {
		favLabel = "🗑 Убрать из избранного"
		add = false
	}
	toggleFavCb := mustCallback(CallbackPayload{
		Action: ActionToggleFav,
		Src:    menu,
		ChatID: chat.ID,
		Add:    &add,
		Page:   gistPage,
	})
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton(favLabel).WithCallbackData(toggleFavCb),
	))

	// Назад
	backMainCb := mustCallback(CallbackPayload{Menu: MenuMain})
	backCb := mustCallback(CallbackPayload{Menu: menu})
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("Домой").WithCallbackData(backMainCb),
		tu.InlineKeyboardButton("← Назад к чатам").WithCallbackData(backCb),
	))

	return tu.InlineKeyboard(rows...)
}

// Меню чатов с пагинацией
func (b *BaseHandler) buildChatsMenu(chats []model.Chat, page int, menu Menu) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Пагинация
	start := page * chatsPerPage
	end := start + chatsPerPage
	if end > len(chats) {
		end = len(chats)
	}

	for i := start; i < end; i++ {
		chat := chats[i]
		label := fmt.Sprintf("📩 %s (%d)", chat.Title, chat.UnreadCount)
		cb := mustCallback(CallbackPayload{Menu: MenuChat, ChatID: chat.ID, Src: menu})
		rows = append(rows, tu.InlineKeyboardRow(
			tu.InlineKeyboardButton(label).WithCallbackData(cb),
		))
	}

	// Кнопки навигации
	var navButtons []telego.InlineKeyboardButton
	if page > 0 {
		cb := mustCallback(CallbackPayload{Menu: menu, Page: page - 1})
		navButtons = append(navButtons, tu.InlineKeyboardButton("◀️").WithCallbackData(cb))
	}
	if end < len(chats) {
		cb := mustCallback(CallbackPayload{Menu: menu, Page: page + 1})
		navButtons = append(navButtons, tu.InlineKeyboardButton("▶️").WithCallbackData(cb))
	}

	if len(navButtons) > 0 {
		rows = append(rows, navButtons)
	}

	// Кнопка назад
	backCb := mustCallback(CallbackPayload{Menu: MenuMain})
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("← Назад").WithCallbackData(backCb),
	))

	return tu.InlineKeyboard(rows...)
}

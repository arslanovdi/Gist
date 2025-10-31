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
	ChatsPerPage = 8
)

// BaseHandler содержит общие зависимости и методы для всех обработчиков
type BaseHandler struct {
	Bot         *telego.Bot
	CoreService CoreService

	LastMessageID int   // id редактируемого сообщения. В боте всегда одно сообщение, которое мы редактируем.
	UserID        int64 // id пользователя = id чата с ним, используется для вывода сообщений ботом.
}

func (b *BaseHandler) showChatDetail(ctx context.Context, chat model.Chat, menu Menu) error {
	log := slog.With("func", "router.showChatDetail")

	inlineKeyboard := b.buildChatDetailMenu(chat.ID, menu, chat.IsFavorite)

	text := fmt.Sprintf("📩 %s\n🔍 Краткий пересказ: %s\n📌 Непрочитано: %d сообщения ", chat.Title, chat.Gist, chat.UnreadCount)

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

// Детали чата
func (b *BaseHandler) buildChatDetailMenu(chatID int64, menu Menu, isFavorite bool) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Действия
	markReadCb := mustCallback(CallbackPayload{Action: ActionMarkRead, ChatID: chatID})
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("✅ Пометить прочитанным").WithCallbackData(markReadCb),
	))

	ttsCb := mustCallback(CallbackPayload{Action: ActionTTS, ChatID: chatID})
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("🔊 Озвучить").WithCallbackData(ttsCb),
	))

	// Кнопка "в избранное" / "убрать из избранного"
	favLabel := "⭐ В избранное"
	add := true
	if isFavorite {
		favLabel = "🗑 Убрать из избранного"
		add = false
	}
	toggleFavCb := mustCallback(CallbackPayload{
		Action: ActionToggleFav,
		Src:    menu,
		ChatID: chatID,
		Add:    &add,
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
	start := page * ChatsPerPage
	end := start + ChatsPerPage
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

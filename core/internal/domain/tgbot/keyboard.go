package tgbot

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

// Главное меню
func (b *Bot) buildMainMenu() *telego.InlineKeyboardMarkup {
	return tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("📬 Непрочитанные чаты").WithCallbackData(mustCallback(CallbackPayload{Menu: MenuUnread})),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("⭐ Избранные чаты").WithCallbackData(mustCallback(CallbackPayload{Menu: MenuFavorites})),
		),
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("⚙️ Настройки").WithCallbackData(mustCallback(CallbackPayload{Menu: MenuSettings})),
		),
	)
}

// Меню чатов с пагинацией
func (b *Bot) buildChatsMenu(chats []model.Chat, page int, menu Menu) *telego.InlineKeyboardMarkup {
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
		cb, _ := CallbackPayload{Menu: MenuChat, ChatID: chat.ID, Src: menu}.String()
		rows = append(rows, tu.InlineKeyboardRow(
			tu.InlineKeyboardButton(label).WithCallbackData(cb),
		))
	}

	// Кнопки навигации
	var navButtons []telego.InlineKeyboardButton
	if page > 0 {
		cb, _ := CallbackPayload{Menu: menu, Page: page - 1}.String()
		navButtons = append(navButtons, tu.InlineKeyboardButton("◀️").WithCallbackData(cb))
	}
	if end < len(chats) {
		cb, _ := CallbackPayload{Menu: menu, Page: page + 1}.String()
		navButtons = append(navButtons, tu.InlineKeyboardButton("▶️").WithCallbackData(cb))
	}

	if len(navButtons) > 0 {
		rows = append(rows, navButtons)
	}

	// Кнопка назад
	backCb, _ := CallbackPayload{Menu: MenuMain}.String()
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("← Назад").WithCallbackData(backCb),
	))

	return tu.InlineKeyboard(rows...)
}

// Детали чата
func (b *Bot) buildChatDetailMenu(chatID int64, menu Menu, isFavorite bool) *telego.InlineKeyboardMarkup {
	var rows [][]telego.InlineKeyboardButton

	// Действия
	markReadCb, _ := CallbackPayload{Action: ActionMarkRead, ChatID: chatID}.String()
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("✅ Пометить прочитанным").WithCallbackData(markReadCb),
	))

	ttsCb, _ := CallbackPayload{Action: ActionTTS, ChatID: chatID}.String()
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
	toggleFavCb, _ := CallbackPayload{
		Action: ActionToggleFav,
		Src:    menu,
		ChatID: chatID,
		Add:    &add,
	}.String()
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton(favLabel).WithCallbackData(toggleFavCb),
	))

	// Назад
	/*	backMenu := MenuUnread
		if menu == MenuFavorites {
			backMenu = MenuFavorites
		}*/
	backCb, _ := CallbackPayload{Menu: menu}.String()
	rows = append(rows, tu.InlineKeyboardRow(
		tu.InlineKeyboardButton("← Назад к чатам").WithCallbackData(backCb),
	))

	return tu.InlineKeyboard(rows...)
}

func mustCallback(cp CallbackPayload) string {
	s, err := cp.String()
	if err != nil {
		panic("callback too long: " + err.Error())
	}
	return s
}

func (b *Bot) showMainMenu(ctx context.Context) {
	log := slog.With("func", "tgbot.showMainMenu")
	log.Debug("showMainMenu")

	inlineKeyboard := b.buildMainMenu()

	if b.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(b.allowedUserID),
			b.LastMessageID,
			"🏠 Главное меню...").WithReplyMarkup(inlineKeyboard)

		_, errE := b.bot.EditMessageText(ctx, message)
		if errE == nil {
			return // Успешно отредактировали
		}
		log.Error("edit message with main menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(b.allowedUserID),
		"🏠 Главное меню...",
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := b.bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with main menu error", slog.Any("error", errS))
		return
	}

	b.LastMessageID = msg.MessageID // Сохраняем номер сообщения
}

func (b *Bot) showUnreadChats(ctx context.Context, chats []model.Chat, page int) {
	log := slog.With("func", "tgbot.showUnreadChats")
	log.Debug("showUnreadChats")

	inlineKeyboard := b.buildChatsMenu(chats, page, MenuUnread)

	if b.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(b.allowedUserID),
			b.LastMessageID,
			fmt.Sprintf("📬 Непрочитанные чаты (%d шт.)", len(chats))).WithReplyMarkup(inlineKeyboard)

		_, errE := b.bot.EditMessageText(ctx, message)
		if errE == nil {
			return // Успешно отредактировали
		}
		log.Error("edit message with unread chats menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(b.allowedUserID),
		fmt.Sprintf("📬 Непрочитанные чаты (%d шт.)", len(chats)),
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := b.bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with unread chats menu error", slog.Any("error", errS))
		return
	}

	b.LastMessageID = msg.MessageID // Сохраняем номер сообщения
}

func (b *Bot) showFavoriteChats(ctx context.Context, chats []model.Chat, page int) {
	log := slog.With("func", "tgbot.showFavoriteChats")
	log.Debug("showFavoriteChats")

	inlineKeyboard := b.buildChatsMenu(chats, page, MenuFavorites)

	if b.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(b.allowedUserID),
			b.LastMessageID,
			fmt.Sprintf("📬 Избранные чаты (%d шт.)", len(chats))).WithReplyMarkup(inlineKeyboard)

		_, errE := b.bot.EditMessageText(ctx, message)
		if errE == nil {
			return // Успешно отредактировали
		}
		log.Error("edit message with favorite chats menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(b.allowedUserID),
		fmt.Sprintf("📬 Избранные чаты (%d шт.)", len(chats)),
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := b.bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with favorite chats menu error", slog.Any("error", errS))
		return
	}

	b.LastMessageID = msg.MessageID // Сохраняем номер сообщения
}

func (b *Bot) showChatDetail(ctx context.Context, chat model.Chat, menu Menu) {
	log := slog.With("func", "tgbot.showChatDetail")
	log.Debug("showChatDetail")

	inlineKeyboard := b.buildChatDetailMenu(chat.ID, menu, chat.IsFavorite)

	text := fmt.Sprintf("📩 %s\n🔍 Краткий пересказ: %s\n📌 Непрочитано: %d сообщения ", chat.Title, chat.Gist, chat.UnreadCount)

	if b.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(b.allowedUserID),
			b.LastMessageID,
			text,
		).WithReplyMarkup(inlineKeyboard)

		_, errE := b.bot.EditMessageText(ctx, message)
		if errE == nil {
			return // Успешно отредактировали
		}
		log.Error("edit message with chat detail menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(b.allowedUserID),
		text,
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := b.bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with chat detail menu error", slog.Any("error", errS))
		return
	}

	b.LastMessageID = msg.MessageID // Сохраняем номер сообщения
}

package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// MainMenuHandler Вывод главного меню
type MainMenuHandler struct {
	*BaseHandler
}

// NewMainMenuHandler конструктор обработчика вывода главного меню.
func NewMainMenuHandler(base *BaseHandler) *MainMenuHandler {
	return &MainMenuHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *MainMenuHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Menu == MenuMain
}

// Handle Реализация интерфейса CallbackHandler
func (h *MainMenuHandler) Handle(ctx *th.Context, query telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "router.MainMenuHandler")
	log.Debug("handling main menu callback")

	// Обязательно сразу отвечаем, что обработчик работает, могут быть проблемы из-за медленных ответов > 10 секунд
	_ = h.Bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))

	return h.showMainMenu(ctx)
}

func (b *BaseHandler) showMainMenu(ctx context.Context) error {
	log := slog.With("func", "tgbot.showMainMenu")
	log.Debug("showMainMenu")

	inlineKeyboard := buildMainMenu()

	if b.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(b.UserID),
			b.LastMessageID,
			"🏠 Главное меню...").WithReplyMarkup(inlineKeyboard)

		_, errE := b.Bot.EditMessageText(ctx, message)
		if errE == nil {
			return nil // Успешно отредактировали
		}
		log.Error("edit message with main menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(b.UserID),
		"🏠 Главное меню...",
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := b.Bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with main menu error", slog.Any("error", errS))
		return fmt.Errorf("send message with main menu error: %w", errS)
	}

	b.LastMessageID = msg.MessageID // Сохраняем номер сообщения
	return nil
}

// Главное меню
func buildMainMenu() *telego.InlineKeyboardMarkup {
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

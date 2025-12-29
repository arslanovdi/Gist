package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// UnreadMenuHandler Вывод списка непрочитанных чатов
type UnreadMenuHandler struct {
	*BaseHandler
}

// NewUnreadMenuHandler конструктор обработчика вывода списка непрочитанных чатов
func NewUnreadMenuHandler(base *BaseHandler) *UnreadMenuHandler {
	return &UnreadMenuHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *UnreadMenuHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Menu == MenuUnread
}

// Handle Реализация интерфейса CallbackHandler
func (h *UnreadMenuHandler) Handle(ctx *th.Context, query telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "router.UnreadMenuHandler")
	log.Debug("handling unread menu callback")

	// Обязательно сразу отвечаем, что обработчик работает, могут быть проблемы из-за медленных ответов > 10 секунд
	_ = h.Bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))

	chats, errF := h.CoreService.GetChatsWithUnreadMessages(ctx)
	if errF != nil {
		log.Error("GetChatsWithUnreadMessages", slog.Any("error", errF))
	}

	return h.showUnreadChats(ctx, chats, payload.Page)
}

func (h *UnreadMenuHandler) showUnreadChats(ctx context.Context, chats []model.Chat, page int) error {
	log := slog.With("func", "router.showUnreadChats")
	log.Debug("showUnreadChats")

	inlineKeyboard := h.buildChatsMenu(chats, page, MenuUnread)

	if h.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(h.UserID),
			h.LastMessageID,
			fmt.Sprintf("📬 Непрочитанные чаты (%d шт.)", len(chats))).WithReplyMarkup(inlineKeyboard)

		_, errE := h.Bot.EditMessageText(ctx, message)
		if errE == nil {
			return nil // Успешно отредактировали
		}
		log.Error("edit message with unread chats menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(h.UserID),
		fmt.Sprintf("📬 Непрочитанные чаты (%d шт.)", len(chats)),
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := h.Bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with unread chats menu error", slog.Any("error", errS))
		return fmt.Errorf("send message with unread chats menu error: %w", errS)
	}

	h.LastMessageID = msg.MessageID // Сохраняем номер сообщения
	return nil
}

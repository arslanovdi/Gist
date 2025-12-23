package router

import (
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// MarkAsReadHandler отметить сообщения чата как прочитанные.
type MarkAsReadHandler struct {
	*BaseHandler
}

// NewMarkAsReadHandler конструктор обработчика помечающего все события выбранного чата как прочитанные.
func NewMarkAsReadHandler(base *BaseHandler) *MarkAsReadHandler {
	return &MarkAsReadHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *MarkAsReadHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionMarkRead
}

// Handle Реализация интерфейса CallbackHandler
func (h *MarkAsReadHandler) Handle(ctx *th.Context, _ telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "router.MarkAsReadHandler")
	log.Debug("handling mark as read callback")

	chat, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID)
	if errD != nil {
		return fmt.Errorf("NewMarkAsReadHandler.GetChatDetail: %w", errD)
	}

	errM := h.CoreService.MarkAsRead(ctx, chat, 0) // 0 отметить все сообщения
	if errM != nil {
		return fmt.Errorf("NewMarkAsReadHandler: %w", errM)
	}

	return nil
}

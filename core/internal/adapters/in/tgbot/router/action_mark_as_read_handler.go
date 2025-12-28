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

	chatDetail, errM := h.CoreService.MarkAsRead(ctx, payload.ChatID, payload.Page)
	if errM != nil {
		return fmt.Errorf("NewMarkAsReadHandler: %w", errM)
	}

	page := 1
	if len(chatDetail.Gist) == 0 { // Может быть при прочтении последнего батча краткого пересказа.
		page = 0
	}

	return h.showChatDetail(ctx, chatDetail, payload.Src, page)
}

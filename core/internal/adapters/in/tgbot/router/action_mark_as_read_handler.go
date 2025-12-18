package router

import (
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
func (h *MarkAsReadHandler) Handle(_ *th.Context, _ telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "router.MarkAsReadHandler")
	log.Debug("handling mark as read callback")

	// TODO implement me
	return nil
}

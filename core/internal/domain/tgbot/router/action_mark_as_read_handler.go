package router

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type MarkAsReadHandler struct {
	*BaseHandler
}

func NewMarkAsReadHandler(base *BaseHandler) *MarkAsReadHandler {
	return &MarkAsReadHandler{BaseHandler: base}
}

func (h *MarkAsReadHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionMarkRead
}

func (h *MarkAsReadHandler) Handle(_ *th.Context, _ telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "router.MarkAsReadHandler")
	log.Debug("handling mark as read callback")

	// TODO implement me
	return nil
}

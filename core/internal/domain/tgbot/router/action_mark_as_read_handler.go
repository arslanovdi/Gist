package router

import (
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
	h.Log.Debug("handling mark as read callback")

	// TODO implement me
	return nil
}

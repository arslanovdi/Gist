package router

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type SettingsMenuHandler struct {
	*BaseHandler
}

func NewSettingsMenuHandler(base *BaseHandler) *SettingsMenuHandler {
	return &SettingsMenuHandler{BaseHandler: base}
}

func (h *SettingsMenuHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Menu == MenuSettings
}

func (h *SettingsMenuHandler) Handle(_ *th.Context, _ telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "router.SettingsMenuHandler")
	log.Debug("handling settings menu callback")
	// TODO implement me
	return nil
}

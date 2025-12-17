package router

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type TTSHandler struct {
	*BaseHandler
}

func NewTTSHandler(base *BaseHandler) *TTSHandler {
	return &TTSHandler{BaseHandler: base}
}

func (h *TTSHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionTTS
}

func (h *TTSHandler) Handle(_ *th.Context, _ telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "TTSHandler")
	log.Debug("handling TTS callback")

	// TODO implement me
	return nil
}

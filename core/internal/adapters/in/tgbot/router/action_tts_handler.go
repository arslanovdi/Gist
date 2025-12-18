package router

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// TTSHandler воспроизвести аудио-пересказ сути чата.
type TTSHandler struct {
	*BaseHandler
}

// NewTTSHandler конструктор обработчика воспроизведения аудио.
func NewTTSHandler(base *BaseHandler) *TTSHandler {
	return &TTSHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *TTSHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionTTS
}

// Handle Реализация интерфейса CallbackHandler
func (h *TTSHandler) Handle(_ *th.Context, _ telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "TTSHandler")
	log.Debug("handling TTS callback")

	// TODO implement me
	return nil
}

package router

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
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
func (h *TTSHandler) Handle(ctx *th.Context, query telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "TTSHandler")
	log.Debug("handling TTS callback")

	// Обязательно сразу отвечаем, что обработчик работает, могут быть проблемы из-за медленных ответов > 10 секунд
	_ = h.Bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))

	// TODO implement me
	return nil
}

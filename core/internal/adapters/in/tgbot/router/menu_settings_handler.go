package router

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// SettingsMenuHandler структура обработчика вывода меню настроек.
type SettingsMenuHandler struct {
	*BaseHandler
}

// NewSettingsMenuHandler конструктор обработчика вывода меню настроек.
func NewSettingsMenuHandler(base *BaseHandler) *SettingsMenuHandler {
	return &SettingsMenuHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *SettingsMenuHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Menu == MenuSettings
}

// Handle Реализация интерфейса CallbackHandler
func (h *SettingsMenuHandler) Handle(ctx *th.Context, query telego.CallbackQuery, _ *CallbackPayload) error {
	log := slog.With("func", "router.SettingsMenuHandler")
	log.Debug("handling settings menu callback")

	// Обязательно сразу отвечаем, что обработчик работает, могут быть проблемы из-за медленных ответов > 10 секунд
	_ = h.Bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))

	// TODO implement me
	return nil
}

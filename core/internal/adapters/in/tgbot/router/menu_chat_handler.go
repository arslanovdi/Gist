package router

import (
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// ChatMenuHandler Вывод информации по выбранному чату
type ChatMenuHandler struct {
	*BaseHandler
}

// NewChatMenuHandler конструктор обработчика вывода информации по выбранному чату.
func NewChatMenuHandler(base *BaseHandler) *ChatMenuHandler {
	return &ChatMenuHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *ChatMenuHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Menu == MenuChat
}

// Handle Реализация интерфейса CallbackHandler
func (h *ChatMenuHandler) Handle(ctx *th.Context, _ telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "router.ChatMenuHandler")
	log.Debug("handling main menu callback")

	chatDetail, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID)
	if errD != nil {
		chatDetail = &model.Chat{}
		log.Error("GetChatDetail", slog.Any("error", errD))
	}

	page := payload.Page
	if page == 0 && len(chatDetail.Gist) > 0 { // Если открываем чат, в котором есть сгенерированный краткий пересказ; page==0 только при первом открытии чата.
		page = 1 // Отображаем первую страницу пересказа.
	}

	return h.showChatDetail(ctx, chatDetail, payload.Src, page) // page меняется по нажатию кнопок вправо/влево
}

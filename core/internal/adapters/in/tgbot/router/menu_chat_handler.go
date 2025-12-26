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
	gistPage := 0

	if payload.Page == 0 { // Если страница не передана, значит краткого пересказа чата еще нет
		_, errG := h.CoreService.GetChatGist(ctx, payload.ChatID) // Получаем краткий пересказ, сохраняем его в кэш.
		if errG != nil {
			log.Error("GetChatGist", slog.Any("error", errG))
		} else {
			gistPage = 1
		}
	} else {
		gistPage = payload.Page
	}

	chatDetail, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID)
	if errD != nil {
		chatDetail = &model.Chat{}
		log.Error("GetChatDetail", slog.Any("error", errD))
	}

	return h.showChatDetail(ctx, *chatDetail, payload.Src, gistPage)
}

package router

import (
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// Вывод информации по выбранному чату
type ChatMenuHandler struct {
	*BaseHandler
}

func NewChatMenuHandler(base *BaseHandler) *ChatMenuHandler {
	return &ChatMenuHandler{BaseHandler: base}
}

func (h *ChatMenuHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Menu == MenuChat
}

func (h *ChatMenuHandler) Handle(ctx *th.Context, _ telego.CallbackQuery, payload *CallbackPayload) error {
	h.Log.Debug("handling main menu callback")

	_, errG := h.CoreService.GetChatGist(ctx, payload.ChatID) // Метод сохраняет суть в структуру Detail
	if errG != nil {
		h.Log.Error("GetChatGist", slog.Any("error", errG))
	}

	chatDetail, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID)
	if errD != nil {
		chatDetail = &model.Chat{}
		h.Log.Error("GetChatDetail", slog.Any("error", errD))
	}

	return h.showChatDetail(ctx, *chatDetail, payload.Src)
}

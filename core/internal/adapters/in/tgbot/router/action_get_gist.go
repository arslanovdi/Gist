package router

import (
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// GistHandler получает краткий пересказ чата.
type GistHandler struct {
	*BaseHandler
}

// NewGistHandler конструктор обработчика кнопки генерации краткого пересказа чата.
func NewGistHandler(base *BaseHandler) *GistHandler {
	return &GistHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *GistHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionGetGist
}

// Handle Реализация интерфейса CallbackHandler
func (h *GistHandler) Handle(ctx *th.Context, _ telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "GistHandler")
	log.Debug("handling get gist callback")

	_, errG := h.CoreService.GetChatGist(ctx, payload.ChatID) // Получаем краткий пересказ, сохраняем его в кэш.
	if errG != nil {
		log.Error("GetChatGist", slog.Any("error", errG))
	}

	chatDetail, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID) // Получаем информацию о чате из кэша
	if errD != nil {
		chatDetail = &model.Chat{}
		log.Error("GetChatDetail", slog.Any("error", errD))
	}

	return h.showChatDetail(ctx, chatDetail, payload.Src, 1) // gistPage после генерации = 1

}

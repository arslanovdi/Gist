package router

import (
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// AddToFavoritesHandler обработчик добавления чата в избранное
type AddToFavoritesHandler struct {
	*BaseHandler
}

// NewAddToFavoritesHandler конструктор обработчика добавления чата в избранное.
func NewAddToFavoritesHandler(base *BaseHandler) *AddToFavoritesHandler {
	return &AddToFavoritesHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *AddToFavoritesHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionToggleFav
}

// Handle Реализация интерфейса CallbackHandler
func (h *AddToFavoritesHandler) Handle(ctx *th.Context, _ telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "router.AddToFavoritesHandler")
	log.Debug("handling add to favorites callback")

	errF := h.CoreService.ChangeFavorites(ctx, payload.ChatID)
	if errF != nil {
		log.Error("ChangeFavorites", slog.Any("error", errF))
	}

	chatDetail, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID)
	if errD != nil {
		chatDetail = &model.Chat{}
		log.Error("GetChatDetail", slog.Any("error", errD))
	}

	return h.showChatDetail(ctx, *chatDetail, payload.Src, payload.Page)
}

package router

import (
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type AddToFavoritesHandler struct {
	*BaseHandler
}

func NewAddToFavoritesHandler(base *BaseHandler) *AddToFavoritesHandler {
	return &AddToFavoritesHandler{BaseHandler: base}
}

func (h *AddToFavoritesHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionToggleFav
}

func (h *AddToFavoritesHandler) Handle(ctx *th.Context, _ telego.CallbackQuery, payload *CallbackPayload) error {
	h.Log.Debug("handling add to favorites callback")

	errF := h.CoreService.ChangeFavorites(ctx, payload.ChatID)
	if errF != nil {
		h.Log.Error("ChangeFavorites", slog.Any("error", errF))
	}

	chatDetail, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID)
	if errD != nil {
		chatDetail = &model.Chat{}
		h.Log.Error("GetChatDetail", slog.Any("error", errD))
	}

	return h.showChatDetail(ctx, *chatDetail, payload.Src)
}

package router

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// FavoritesMenuHandler Вывод списка избранных чатов.
type FavoritesMenuHandler struct {
	*BaseHandler
}

func NewFavoritesMenuHandler(base *BaseHandler) *FavoritesMenuHandler {
	return &FavoritesMenuHandler{BaseHandler: base}
}

func (h *FavoritesMenuHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Menu == MenuFavorites
}

func (h *FavoritesMenuHandler) Handle(ctx *th.Context, _ telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "router.FavoritesMenuHandler")
	log.Debug("handling favorites menu callback")

	chats, errF := h.CoreService.GetFavoriteChats(ctx)
	if errF != nil {
		log.Error("GetFavoriteChats", slog.Any("error", errF))
	}

	return h.showFavoriteChats(ctx, chats, payload.Page)
}

func (h *FavoritesMenuHandler) showFavoriteChats(ctx context.Context, chats []model.Chat, page int) error {
	log := slog.With("func", "router.showFavoriteChats")
	log.Debug("showFavoriteChats")

	inlineKeyboard := h.buildChatsMenu(chats, page, MenuFavorites)

	if h.LastMessageID != 0 {
		// Пытаемся отредактировать
		message := tu.EditMessageText(
			tu.ID(h.UserID),
			h.LastMessageID,
			fmt.Sprintf("📬 Избранные чаты (%d шт.)", len(chats))).WithReplyMarkup(inlineKeyboard)

		_, errE := h.Bot.EditMessageText(ctx, message)
		if errE == nil {
			return nil // Успешно отредактировали
		}
		log.Error("edit message with favorite chats menu error", slog.Any("error", errE))
		// Иначе — отправим новое
	}

	// Отправляем новое
	message := tu.Message(
		tu.ID(h.UserID),
		fmt.Sprintf("📬 Избранные чаты (%d шт.)", len(chats)),
	).WithReplyMarkup(inlineKeyboard)

	msg, errS := h.Bot.SendMessage(ctx, message)
	if errS != nil {
		log.Error("send message with favorite chats menu error", slog.Any("error", errS))
		return fmt.Errorf("send message with favorite chats menu error: %w", errS)
	}

	h.LastMessageID = msg.MessageID // Сохраняем номер сообщения
	return nil
}

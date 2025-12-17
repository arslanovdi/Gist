package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

const MaxDataSize = 64 // 64 байта callback payload это ограничение Telegram

type Menu int8

const (
	MenuMain Menu = iota + 1
	MenuUnread
	MenuFavorites
	MenuSettings
	MenuChat
)

type Action int8

const (
	ActionMarkRead  Action = iota + 1 // ✅ Пометить прочитанным
	ActionTTS                         // 🔊 Озвучить"
	ActionToggleFav                   // ⭐ В избранное; 🗑 Убрать из избранного
)

// CallbackPayload — данные, сериализуемые в callback_data
type CallbackPayload struct {
	Menu   Menu   `json:"m,omitempty"`   // MenuMain, MenuUnread, MenuFavorites, MenuChat, MenuSettings	 	int8
	Page   int    `json:"p,omitempty"`   // номер страницы, при выводе списка чатов.						int
	ChatID int64  `json:"c,omitempty"`   // ID чата	требуется при выводе инлан-кнопок со списком чатов		int64
	Src    Menu   `json:"s,omitempty"`   // MenuUnread или MenuFavorites. тип списка чатов					int8
	Action Action `json:"a,omitempty"`   // ActionMarkRead, ActionTTS, ActionToggleFav, и т.д.				int8
	Add    *bool  `json:"add,omitempty"` // для ActionToggleFav												bool
}

// Сериализация в callback_data (до 64 байт)
func (cp CallbackPayload) String() (string, error) {
	data, err := json.Marshal(cp)
	if err != nil {
		return "", err
	}
	if len(data) > MaxDataSize {
		return "", fmt.Errorf("callback_data too long: %d bytes", len(data))
	}
	return string(data), nil
}

// parseCallback
func parseCallback(data string) (*CallbackPayload, error) {
	var cp CallbackPayload
	if err := json.Unmarshal([]byte(data), &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

func mustCallback(cp CallbackPayload) string { // TODO сделать тесты на максимальный размер payload
	s, err := cp.String()
	if err != nil {
		slog.With("func", "router.mustCallback").Error("callback serialization failed:", slog.Any("error", err))
		panic("callback too long: " + err.Error())
	}
	return s
}

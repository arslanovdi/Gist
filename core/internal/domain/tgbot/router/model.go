package router

import (
	"encoding/json"
	"fmt"
)

type Menu string

const (
	MenuMain      Menu = "main"
	MenuUnread    Menu = "unread"
	MenuFavorites Menu = "favorites"
	MenuSettings  Menu = "settings"
	MenuChat      Menu = "chat"
)

type Action string

const (
	ActionMarkRead  Action = "mark_read"  // ✅ Пометить прочитанным
	ActionTTS       Action = "tts"        // 🔊 Озвучить"
	ActionToggleFav Action = "toggle_fav" // ⭐ В избранное; 🗑 Убрать из избранного
)

// CallbackPayload — данные, сериализуемые в callback_data
type CallbackPayload struct {
	Menu   Menu   `json:"m,omitempty"`   // MenuMain, MenuUnread, MenuFavorites, MenuChat, MenuSettings
	Page   int    `json:"p,omitempty"`   // номер страницы, при выводе списка чатов.
	ChatID int64  `json:"c,omitempty"`   // ID чата	требуется при выводе инлан-кнопок со списком чатов
	Src    Menu   `json:"s,omitempty"`   // MenuUnread или MenuFavorites. тип списка чатов
	Action Action `json:"a,omitempty"`   // ActionMarkRead, ActionTTS, ActionToggleFav, и т.д.
	Add    *bool  `json:"add,omitempty"` // для ActionToggleFav
}

// Сериализация в callback_data (до 64 байт)
func (cp CallbackPayload) String() (string, error) {
	data, err := json.Marshal(cp)
	if err != nil {
		return "", err
	}
	if len(data) > 64 {
		return "", fmt.Errorf("callback_data too long: %d bytes", len(data))
	}
	return string(data), nil
}

// ParseCallback
func ParseCallback(data string) (*CallbackPayload, error) {
	var cp CallbackPayload
	if err := json.Unmarshal([]byte(data), &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

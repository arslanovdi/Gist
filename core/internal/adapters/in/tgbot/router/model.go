package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

const maxDataSize = 64 // 64 байта callback payload это ограничение Telegram

// Menu тип меню в Telegram боте.
// Используется для навигации между разделами интерфейса.
type Menu int8

// Список вариантов меню
const (
	MenuMain      Menu = iota + 1 // Главное меню
	MenuUnread                    // Список чатов с непрочитанными сообщениями
	MenuFavorites                 // Список избранных чатов
	MenuSettings                  // Меню настроек
	MenuChat                      // Меню выбранного чата
)

// Action тип действия, которое может быть выполнено с чатом Telegram.
type Action int8

// Список вариантов действий
const (
	ActionMarkRead  Action = iota + 1 // ✅ Пометить прочитанным
	ActionTTS                         // 🔊 Озвучить"
	ActionToggleFav                   // ⭐ В избранное; 🗑 Убрать из избранного
)

// CallbackPayload — данные, сериализуемые в callback_data
type CallbackPayload struct {
	Menu   Menu   `json:"m,omitempty"`   // MenuMain, MenuUnread, MenuFavorites, MenuChat, MenuSettings	 	int8
	Page   int    `json:"p,omitempty"`   // Номер страницы, при выводе списка чатов.
	ChatID int64  `json:"c,omitempty"`   // ID чата	требуется при выводе инлайн-кнопок со списком чатов
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
	if len(data) > maxDataSize {
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

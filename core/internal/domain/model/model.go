// Package model содержит общие для разных слоев типы, ошибки.
package model

import (
	"github.com/gotd/td/tg"
)

// Chat структура телеграмм чата
type Chat struct {
	Title             string      // From Chats.Title
	ID                int64       // From Chats.ID
	UnreadCount       int         // From Dialogs.UnreadCount
	IsFavorite        bool        // TODO Поле Временно, вынести настройки в БД
	Gist              []BatchGist // Краткий пересказ каждого батча сообщений, батчи формируются в соответствии с контекстным окном LLM.
	Peer              tg.InputPeerClass
	LastReadMessageID int
}

// BatchGist структура хранит краткий пересказ батча сообщений
type BatchGist struct {
	LastMessageID int // ID последнего сообщения, в данном батче
	MessageCount  int
	Gist          string // Краткий пересказ
}

// Package model содержит общие для разных слоев типы, ошибки.
package model

import (
	"time"

	"github.com/gotd/td/tg"
)

// Chat структура телеграмм чата
type Chat struct {
	Title             string      // From Chats.Title
	ID                int64       // From Chats.ID
	UnreadCount       int         // From Dialogs.UnreadCount
	Skipped           int         // Кол-во пропущенных сообщений (сообщения без текста фото и т.п.)
	IsFavorite        bool        // TODO Поле Временно, вынести настройки в БД
	Gist              []BatchGist // Краткий пересказ каждого батча сообщений, батчи формируются в соответствии с контекстным окном LLM.
	AudioFile         string      // Имя файла с аудиопересказом всех батчей.
	Peer              tg.InputPeerClass
	LastReadMessageID int
}

// BatchGist структура хранит краткий пересказ батча сообщений
type BatchGist struct {
	LastMessageID   int       // ID последнего сообщения, в данном батче
	LastMessageData time.Time // Метка времени последнего сообщения
	MessageCount    int
	Gist            string // Краткий пересказ
	AudioFile       string // Имя файла с аудиопересказом батча
}

type AudioGist struct {
	AudioFile string
	Caption   string
}

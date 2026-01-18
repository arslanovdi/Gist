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
	Audio             []AudioGist // Описание файла(ов) с аудиопересказом всех батчей. Файлов может быть несколько, если размер превышает максимально разрешенный.
	Peer              tg.InputPeerClass
	LastReadMessageID int

	Messages []Message
}

// BatchGist структура хранит краткий пересказ батча сообщений
type BatchGist struct {
	FirstMessageData time.Time // Метка времени первого сообщения
	LastMessageID    int       // ID последнего сообщения, в данном батче
	LastMessageData  time.Time // Метка времени последнего сообщения
	MessageCount     int
	Gist             string      // Краткий пересказ
	Audio            []AudioGist // Предполагается, что аудиопересказ одного батча хранится в одном файле. Вероятность того, что аудиопересказ будет больше 50 Мб есть, но стремится к нулю.
}

type AudioGist struct {
	AudioFile string // Путь к файлу
	Caption   string // Описание, выводимое в голосовом сообщении телеграмм бота.
}

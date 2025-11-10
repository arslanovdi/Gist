package model

import (
	"fmt"
	"time"
)

type Message struct {
	ID           int       // ID сообщения в Telegram
	ChatID       int64     // ID чата, откуда пришло сообщение
	SenderID     int64     // ID отправителя
	Username     string    // Никнейм отправителя (может быть пустым, если скрыт)
	FirstName    string    // Имя отправителя
	LastName     string    // Фамилия отправителя (может быть пустой)
	Text         string    // Текст сообщения
	Timestamp    time.Time // Время отправки сообщения
	IsEdited     bool      // Было ли сообщение отредактировано
	MessageType  string    // Тип сообщения (text, photo, document и т.д.)
	MediaURL     string    // URL медиафайла, если есть
	ReplyToMsgID int64     // ID сообщения, на которое отвечают (0, если не ответ)
	IsForwarded  bool      // Является ли сообщение пересланным
}

// FormatForAnalysis возвращает отформатированное сообщение для анализа
func (m *Message) FormatForAnalysis() string {
	return fmt.Sprintf("[%s] %s: %s",
		m.Timestamp.Format("2006-01-02 15:04:05"),
		m.Username,
		m.Text,
	)
}

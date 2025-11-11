package model

import (
	"fmt"
	"time"
)

type Message struct {
	ID           int       // ID сообщения в Telegram
	SenderID     int64     // ID отправителя
	Text         string    // Текст сообщения
	Timestamp    time.Time // Время отправки сообщения
	IsEdited     bool      // Было ли сообщение отредактировано
	ReplyToMsgID int       // ID сообщения, на которое отвечают (0, если не ответ)
	IsForwarded  bool      // Является ли сообщение пересланным
}

// FormatForAnalysis возвращает отформатированное сообщение для анализа
func (m *Message) FormatForAnalysis() string {
	return fmt.Sprintf("[%s] %d: %s",
		m.Timestamp.Format("2006-01-02 15:04:05"),
		m.SenderID,
		m.Text,
	)
}

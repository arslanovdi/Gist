package model

import (
	"fmt"
	"time"
)

type Message struct {
	ID           int       `json:"id"`              // ID сообщения в Telegram
	SenderID     int64     `json:"sender_id"`       // ID отправителя
	Text         string    `json:"text"`            // Текст сообщения
	Timestamp    time.Time `json:"timestamp"`       // Время отправки сообщения
	IsEdited     bool      `json:"is_edited"`       // Было ли сообщение отредактировано
	ReplyToMsgID int       `json:"reply_to_msg_id"` // ID сообщения, на которое отвечают (0, если не ответ)
	IsForwarded  bool      `json:"is_forwarded"`    // Является ли сообщение пересланным
}

// FormatForAnalysis возвращает отформатированное сообщение для анализа
func (m *Message) FormatForAnalysis() string {
	return fmt.Sprintf("[%s] %d: %s",
		m.Timestamp.Format("2006-01-02 15:04:05"),
		m.SenderID,
		m.Text,
	)
}

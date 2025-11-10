package model

import "github.com/gotd/td/tg"

type Credential struct {
	Phone string
	Code  string
}

type Chat struct {
	Title             string // From Chats.Title
	ID                int64  // From Chats.ID
	UnreadCount       int    // From Dialogs.UnreadCount
	IsFavorite        bool   // TODO Поле Временно, вынести настройки в БД
	Gist              string
	Peer              tg.InputPeerClass
	LastReadMessageID int
}

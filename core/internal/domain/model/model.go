package model

type Credential struct {
	Phone string
	Code  string
}

type UserState string

const (
	AuthGetPhone = UserState("аутентификация, ввод номера телефона")
	AuthGetCode  = UserState("аутентификация, ввод кода подтверждения")
	AuthDone     = UserState("аутентификация, данные получены")
)

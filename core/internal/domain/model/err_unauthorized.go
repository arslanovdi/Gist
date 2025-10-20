package model

// ErrUnauthorized ошибка авторизации телеграм клиента
type ErrUnauthorized struct {
	Phone     chan<- string // Канал для передачи номера телефона
	Code      chan<- string // Канал для передачи кода подтверждения
	AuthError chan error    // Канал вернет ошибку при неуспешной авторизации или nil если авторизация пройдена.
}

func (e *ErrUnauthorized) Error() string {
	return "Unauthorized client"
}

func NewErrUnauthorized(phone, code chan<- string, authError chan error) error {
	return &ErrUnauthorized{
		Phone:     phone,
		Code:      code,
		AuthError: authError,
	}
}

package model

import "errors"

// ErrNotReady Ошибка готовности телеграмм клиента
var ErrNotReady = errors.New("client not ready")

// ErrChatNotFoundInCache Чат в кэше не найден
var ErrChatNotFoundInCache = errors.New("chat not found in cache")

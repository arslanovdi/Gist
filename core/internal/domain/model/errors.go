package model

import "errors"

var ErrNotReady = errors.New("client not ready")
var ErrChatNotFoundInCache = errors.New("chat not found in cache")

package tgclient

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// Authenticate выполняет аутентификацию пользователя в Telegram API.
//
// Запрашивает код подтверждения через консольный ввод
func (s *Session) Authenticate(ctx context.Context) error {
	log := slog.With("func", "tgclient.authenticate", slog.Any("user_id", s.userID))

	// Функция для запроса кода подтверждения в консоли
	codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		fmt.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(code), nil
	}

	// Аутентификация без пароля
	flow := auth.NewFlow(
		auth.CodeOnly(s.phone, auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)

	errF := flow.Run(ctx, s.client.Auth())
	if errF != nil {
		// При ошибке аутентификации удаляем сессию и пробуем снова
		if errR := os.Remove("session.json"); errR != nil && !os.IsNotExist(errR) {
			log.Error("Warning: failed to remove session file", slog.Any("error", errR))
		}
		log.Debug("authentication failed", slog.Any("error", errF), slog.Int64("user_id", s.userID))
		return fmt.Errorf("authentication failed: %w", errF)
	}

	log.Debug("Authentication successful!", slog.Int64("user_id", s.userID))

	return nil
}

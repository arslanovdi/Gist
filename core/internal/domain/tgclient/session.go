package tgclient

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type Session struct {
	UserID     int64
	Cred       model.Credential // Номер телефона, код подтверждения.
	Client     *telegram.Client
	LastAccess time.Time
	CreatedAt  time.Time
}

func (s *Session) GetAllChats(ctx context.Context) ([]string, error) {
	log := slog.With("func", "user.GetAllChats")

	chats := make([]string, 0)

	// Запуск клиента.
	if err := s.Client.Run(ctx, func(ctx context.Context) error {
		// Проверяем, авторизованы ли мы уже
		authStatus, err := s.Client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("get auth status failed: %w", err)
		}

		// Если не авторизованы, выполняем полный процесс авторизации
		if !authStatus.Authorized {
			log.Debug("Not authenticated, starting authentication flow...", slog.Int64("user_id", s.UserID))
			if errA := s.Authenticate(ctx); errA != nil {
				return errA
			}
		} else {
			log.Debug("Already authenticated, using existing session...", slog.Int64("user_id", s.UserID))
		}

		// После успешной аутентификации получаем список чатов
		api := s.Client.API()

		// Получение списка диалогов, отсортированных по дате последнего сообщения...
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			Limit:      100,                  // Официальный лимит 100-200 диалогов за запрос.
			OffsetDate: 0,                    // lastMessageDate, // Из последнего полученного сообщения
			OffsetID:   0,                    // lastMessageID, // Из последнего полученного сообщения
			OffsetPeer: &tg.InputPeerEmpty{}, // lastPeer Из последнего диалога

		})
		if err != nil {
			return fmt.Errorf("get dialogs error: %v", err)
		}

		// Обработка результатов, должны быть готовы к обработке всех типов ответа.
		switch d := dialogs.(type) {
		case *tg.MessagesDialogs: // https://core.telegram.org/constructor/messages.dialogs Это полный список диалогов, выдается если умещается в один ответ сервера.
			log.Info("MessagesDialogs")
		case *tg.MessagesDialogsSlice: // https://core.telegram.org/constructor/messages.dialogsSlice	часть диалогов (страница срез/пагинация)
			log.Info("MessagesDialogsSlice")
			log.Debug("dialogs count", d.Count)
		case *tg.MessagesDialogsNotModified: // https://core.telegram.org/constructor/messages.dialogsNotModified уведомление, что со времени последнего запроса список диалогов не изменился. Возвращается, если при вызове MessageGetDialogs передать hash.
			log.Info("MessagesDialogsNotModified")
		default:
			log.Error("Unexpected response type")
		}

		return nil
	}); err != nil {
		log.Error("Client error", slog.Any("error", err))
		return nil, err
	}

	return chats, nil
}

func (s *Session) UpdateLastAccess() {
	s.LastAccess = time.Now()
}

func (s *Session) IsExpired(ttl time.Duration) bool {
	return time.Since(s.LastAccess) > ttl
}

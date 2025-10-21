package tgclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gotd/td/tg"
)

// GetAllChats Возвращает список всех чатов
func (s *Session) GetAllChats(ctx context.Context) ([]string, error) {
	log := slog.With("func", "tgclient.GetAllChats")

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

		// Получаем список чатов
		api := s.Client.API()

		// Получение списка диалогов, отсортированных по дате последнего сообщения...
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			Limit:      100,                  // Официальный лимит 100-200 диалогов за запрос.
			OffsetDate: 0,                    // lastMessageDate, // Из последнего полученного сообщения
			OffsetID:   0,                    // lastMessageID, // Из последнего полученного сообщения
			OffsetPeer: &tg.InputPeerEmpty{}, // lastPeer Из последнего диалога

		})
		if err != nil {
			return fmt.Errorf("get dialogs error: %w", err)
		}

		// Обработка результатов, должны быть готовы к обработке всех типов ответа. TODO если запрашиваем с лимитом ответ всегда будет MessagesDialogsSlice?
		switch d := dialogs.(type) {
		case *tg.MessagesDialogs: // https://core.telegram.org/constructor/messages.dialogs Это полный список диалогов, выдается если умещается в один ответ сервера.
			log.Info("MessagesDialogs")
		case *tg.MessagesDialogsSlice: // https://core.telegram.org/constructor/messages.dialogsSlice	часть диалогов (страница срез/пагинация)
			log.Info("MessagesDialogsSlice")
			log.Debug("dialogs", slog.Int("count", d.Count))
			chats = GetChatsList(d)

		case *tg.MessagesDialogsNotModified: // https://core.teleram.org/constructor/messages.dialogsNotModified уведомление, что со времени последнего запроса список диалогов не изменился. Возвращается, если при вызове MessageGetDialogs передать hash.
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

func GetChatsList(dialogs *tg.MessagesDialogsSlice) []string {
	log := slog.With("func", "tgclient.GetChatsList")
	log.Debug("MessagesDialogsSlice chats", slog.Int("count", len(dialogs.Chats)))

	chats := make([]string, 0)
	for _, chat := range dialogs.Chats {
		switch c := chat.(type) {
		case *tg.Chat:
			chats = append(chats, fmt.Sprintf("👥 Group: %s (ID: %d)\n", c.Title, c.ID))
		case *tg.Channel:
			if c.Broadcast {
				chats = append(chats, fmt.Sprintf("📢 Channel: %s (ID: %d)\n", c.Title, c.ID))
			} else {
				chats = append(chats, fmt.Sprintf("💬 Supergroup: %s (ID: %d)\n", c.Title, c.ID))
			}
		case *tg.ChatForbidden:
			chats = append(chats, fmt.Sprintf("🚫 Forbidden chat: %s (ID: %d)\n", c.Title, c.ID))
		case *tg.ChannelForbidden:
			chats = append(chats, fmt.Sprintf("🚫 Forbidden channel: %s (ID: %d)\n", c.Title, c.ID))
		default:
			chats = append(chats, fmt.Sprintf("❓ Unknown chat type: %T\n", c))
		}
	}
	return chats
}

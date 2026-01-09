package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// MarkAsRead отметить сообщения чата как прочитанные. Нумерация страниц начинается с 1.
func (g *Gist) MarkAsRead(ctx context.Context, chatID int64, pageID int) (*model.Chat, error) {

	chat, errD := g.GetChatDetail(ctx, chatID)
	if errD != nil {
		return nil, fmt.Errorf("core.MarkAsRead: %w", errD)
	}

	lastMessageID := 0 // Если страница не задана == 0, отмечаем прочитанными ВСЕ сообщения чата.
	if pageID > 0 {    // Иначе отмечаем прочитанными только сообщения до текущего батча с кратким пересказом.
		lastMessageID = chat.Gist[pageID-1].LastMessageID
	}

	errM := g.tgClient.MarkAsRead(ctx, chat, lastMessageID)
	if errM != nil {
		return nil, fmt.Errorf("core.MarkAsRead: %w", errM)
	}

	chat.LastReadMessageID = lastMessageID // обновляем ID последнего прочитанного сообщения

	if pageID > 0 {
		for i := 0; i < pageID; i++ { // Очистка всех пересказов до текущего включительно.
			chat.UnreadCount -= chat.Gist[i].MessageCount // уменьшаем количество непрочитанных сообщений в чате
			deleteFile(chat.Gist[i].AudioFile)            // удаляем файл с аудиопересказом, если есть
		}
		chat.Gist = slices.Delete(chat.Gist, 0, pageID) // удаляем батчи с пересказом
		deleteFile(chat.AudioFile)                      // удаляем файл с полным аудиопересказом, если есть
		chat.AudioFile = ""                             // Обнуляем полный аудиопересказ, т.к. часть пометили прочитанным

		return chat, nil
	}

	for i := 0; i < len(chat.Gist); i++ {
		deleteFile(chat.Gist[i].AudioFile) // удаляем файл с аудиопересказом, если есть
	}

	chat.UnreadCount = 0 // количество непрочитанных сообщений в чате = 0
	chat.Gist = nil      // удалили все краткие пересказы

	deleteFile(chat.AudioFile) // удаляем файл с полным аудиопересказом, если есть
	chat.AudioFile = ""

	return chat, nil
}

func deleteFile(name string) {
	log := slog.With("func", "core.deleteFile")

	if name == "" {
		return
	}

	errR := os.Remove(name)
	if errR != nil {
		log.Error("error removing file", errR, slog.String("name", name))
	}

}

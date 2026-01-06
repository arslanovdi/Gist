package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arslanovdi/Gist/core/internal/infra/ffmpeg"
)

// GetAudioGist возвращает имя файла с аудиопересказом
// batchID - номер батча, для которого нужно вернуть аудиопересказ, если batchID = 0 возвращаем аудиопересказ всего чата	todo потестить режимы.
func (g *Gist) GetAudioGist(ctx context.Context, chatID int64, batchID int) (string, error) {

	log := slog.With("func", "core.GetAudioGist")
	log.Debug("start GetAudioGist")

	chat, errD := g.GetChatDetail(ctx, chatID)
	if errD != nil {
		return "", fmt.Errorf("get chat detail: %w", errD)
	}

	// Если batchGist пустой, возвращаем ошибку
	if len(chat.Gist) == 0 {
		return "", fmt.Errorf("batchGist is empty")
	}

	// Запросили аудиопересказ батча
	if batchID > 0 {
		if batchID > len(chat.Gist) {
			return "", fmt.Errorf("batch ID %d exceeds available batches count (%d)", batchID, len(chat.Gist))
		}

		if chat.Gist[batchID-1].AudioFile != "" {
			return chat.Gist[batchID-1].AudioFile, nil // возвращаем файл с нужным аудиопересказом
		}

		// генерируем аудиопересказ, сохраняется в chat по указателю
		errG := g.ttsClient.GenerateAudioGist(ctx, chat)
		if errG != nil {
			return "", fmt.Errorf("core.GetAudioGist generate error: %w", errG)
		}

		if chat.Gist[batchID-1].AudioFile != "" {
			return chat.Gist[batchID-1].AudioFile, nil // возвращаем файл с нужным аудиопересказом
		}

		return "", fmt.Errorf("core.GetAudioGist unknown error, batchID:%d", batchID)
	}

	// запросили полный аудиопересказ
	if chat.AudioFile != "" {
		return chat.AudioFile, nil
	}

	// проверка наличия аудиопересказов батчей
	for i := range chat.Gist {
		if chat.Gist[i].AudioFile == "" {
			log.Debug("Нет аудиопересказа батча", slog.Int("batch index", i))

			// генерируем аудиопересказы
			errG := g.ttsClient.GenerateAudioGist(ctx, chat)
			if errG != nil {
				return "", fmt.Errorf("core.GetAudioGist generate error: %w", errG)
			}
			break
		}
	}

	list := make([]string, 0)
	for i := range chat.Gist {
		list = append(list, chat.Gist[i].AudioFile)
	}

	audioFile := chat.Title + ".mp3"
	// собираем полный аудиопересказ из батчей
	errC := ffmpeg.ConcatMP3(list, audioFile)
	if errC != nil {
		return "", fmt.Errorf("core.GetAudioGist ffmpeg concatMP3: %w", errC)
	}

	chat.AudioFile = audioFile // сохраняем имя файла

	return audioFile, nil
}

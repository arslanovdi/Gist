package core

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/ffmpeg"
)

// GetAudioGist возвращает имя файла с аудиопересказом
// batchID - номер батча, для которого нужно вернуть аудиопересказ, если batchID = 0 возвращаем аудиопересказ всего чата	todo потестить режимы.
func (g *Gist) GetAudioGist(ctx context.Context, chatID int64, batchID int) (*model.AudioGist, error) {

	log := slog.With("func", "core.GetAudioGist")
	log.Debug("start GetAudioGist")

	chat, errD := g.GetChatDetail(ctx, chatID)
	if errD != nil {
		return nil, fmt.Errorf("get chat detail: %w", errD)
	}

	// Если batchGist пустой, возвращаем ошибку
	if len(chat.Gist) == 0 {
		return nil, fmt.Errorf("batchGist is empty")
	}

	// Запросили аудиопересказ батча
	if batchID > 0 {
		if batchID > len(chat.Gist) {
			return nil, fmt.Errorf("batch ID %d exceeds available batches count (%d)", batchID, len(chat.Gist))
		}

		caption := fmt.Sprintf("%s до %s", chat.Title, chat.Gist[batchID-1].LastMessageData)

		if chat.Gist[batchID-1].AudioFile != "" {
			return &model.AudioGist{
				AudioFile: chat.Gist[batchID-1].AudioFile,
				Caption:   caption,
			}, nil // возвращаем файл с нужным аудиопересказом
		}

		// генерируем аудиопересказ, сохраняется в chat по указателю
		errG := g.llmClient.GenerateAudioGist(ctx, chat, batchID)
		if errG != nil {
			return nil, fmt.Errorf("core.GetAudioGist generate error: %w", errG)
		}

		if chat.Gist[batchID-1].AudioFile != "" {
			return &model.AudioGist{
				AudioFile: chat.Gist[batchID-1].AudioFile,
				Caption:   caption,
			}, nil // возвращаем файл с нужным аудиопересказом
		}

		return nil, fmt.Errorf("core.GetAudioGist unknown error, batchID:%d", batchID)
	}

	// запросили полный аудиопересказ
	if chat.AudioFile != "" {
		return &model.AudioGist{
			AudioFile: chat.AudioFile,
			Caption:   fmt.Sprintf("%s полный пересказ до %s", chat.Title, chat.Gist[len(chat.Gist)-1].LastMessageData),
		}, nil
	}

	// проверка наличия аудиопересказов батчей
	for i := range chat.Gist {
		if chat.Gist[i].AudioFile == "" {
			log.Debug("Нет аудиопересказа батча", slog.Int("batch index", i))

			// генерируем аудиопересказы
			errG := g.llmClient.GenerateAudioGist(ctx, chat, 0)
			if errG != nil {
				return nil, fmt.Errorf("core.GetAudioGist generate error: %w", errG)
			}
			break
		}
	}

	list := make([]string, 0)
	for i := range chat.Gist {
		list = append(list, chat.Gist[i].AudioFile)
	}

	audioFile := filepath.Join(g.cfg.Project.AudioPath, fmt.Sprintf("%d.mp3", chat.ID))

	// собираем полный аудиопересказ из батчей
	errC := ffmpeg.ConcatMP3(list, audioFile)
	if errC != nil {
		return nil, fmt.Errorf("core.GetAudioGist ffmpeg concatMP3: %w", errC)
	}

	chat.AudioFile = audioFile // сохраняем имя файла

	return &model.AudioGist{
		AudioFile: audioFile,
		Caption:   fmt.Sprintf("%s полный пересказ до %s", chat.Title, chat.Gist[len(chat.Gist)-1].LastMessageData),
	}, nil
}

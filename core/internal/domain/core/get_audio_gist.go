package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/ffmpeg"
	"github.com/arslanovdi/Gist/core/internal/infra/utils"
)

// GetAudioGist возвращает имя файла с аудиопересказом
// batchID - номер батча, для которого нужно вернуть аудиопересказ, если batchID = 0 возвращаем аудиопересказ всего чата	todo потестить режимы.
func (g *Gist) GetAudioGist(ctx context.Context, chatID int64, batchID int) ([]model.AudioGist, error) {

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

		if len(chat.Gist[batchID-1].Audio) > 0 { // Аудио уже сгенерировано
			return chat.Gist[batchID-1].Audio, nil // возвращаем нужный аудиопересказ
		}

		// Генерируем аудиопересказ, сохраняется в chat по указателю
		errG := g.llmClient.GenerateAudioGist(ctx, chat, batchID)
		if errG != nil {
			return nil, fmt.Errorf("core.GetAudioGist generate error: %w", errG)
		}

		return chat.Gist[batchID-1].Audio, nil // возвращаем файл с нужным аудиопересказом
	}

	// запросили полный аудиопересказ
	if len(chat.Audio) > 0 { // Он уже есть
		return chat.Audio, nil
	}

	// проверка наличия аудиопересказов батчей
	for i := range chat.Gist {
		if len(chat.Gist[i].Audio) == 0 {
			log.Debug("Нет аудиопересказа батча", slog.Int("batch index", i))

			// Генерируем аудиопересказы, при batchID = 0 сгенерируются все отсутствующие
			errG := g.llmClient.GenerateAudioGist(ctx, chat, 0)
			if errG != nil {
				return nil, fmt.Errorf("core.GetAudioGist generate error: %w", errG)
			}
			break
		}
	}

	list := make([]model.AudioGist, 0) // Собираем список файлов
	for i := range chat.Gist {
		for j := range chat.Gist[i].Audio {
			list = append(list, chat.Gist[i].Audio[j])
		}
	}

	audioFile := filepath.Join(g.cfg.Project.AudioPath, fmt.Sprintf("%d.mp3", chat.ID))

	// собираем полный аудиопересказ из батчей
	errC := ffmpeg.ConcatMP3(list, audioFile)
	if errC != nil {
		return nil, fmt.Errorf("core.GetAudioGist ffmpeg concatMP3: %w", errC)
	}

	info, errS := os.Stat(audioFile)
	if errS != nil { // ошибка, возвращаем полный файл
		log.Error("get FileInfo of audio file error", slog.String("filename", audioFile), slog.Any("error", errS))
		chat.Audio = []model.AudioGist{
			{
				AudioFile: audioFile,
				Caption: fmt.Sprintf("%s (%s)\nПолный пересказ от %s",
					chat.Title,
					utils.FormatDurationShort(chat.Gist[len(chat.Gist)-1].LastMessageData.Sub(chat.Gist[0].FirstMessageData)),
					utils.FormatDateShort(chat.Gist[0].FirstMessageData),
				),
			},
		}

		return chat.Audio, nil // Ошибку получения информации о файле логирую и игнорирую.
	}

	if info.Size() > g.cfg.LLM.TTS.MaxAudioFileSize*1024*1024 { // Размер файла превышает максимально разрешенный
		files, errT := ffmpeg.SplitMP3(audioFile, g.cfg.LLM.TTS.MaxAudioFileSize) // Разбиваем на несколько
		if errT != nil {
			log.Error("Trim audiofile error", slog.String("filename", audioFile), slog.Any("error", errT))
		}

		for index := range files { // добавляем файлы
			chat.Audio = append(chat.Audio, model.AudioGist{
				AudioFile: files[index].AudioFile,
				Caption: fmt.Sprintf("%s (%s)\nПолный пересказ от %s\npart %d",
					chat.Title,
					utils.FormatDurationShort(chat.Gist[len(chat.Gist)-1].LastMessageData.Sub(chat.Gist[0].FirstMessageData)),
					utils.FormatDateShort(chat.Gist[0].FirstMessageData),
					index+1,
				),
			})
		}
	} else { // иначе добавляем один файл
		chat.Audio = append(chat.Audio, model.AudioGist{
			AudioFile: audioFile,
			Caption: fmt.Sprintf("%s (%s)\nПолный пересказ от %s",
				chat.Title,
				utils.FormatDurationShort(chat.Gist[len(chat.Gist)-1].LastMessageData.Sub(chat.Gist[0].FirstMessageData)),
				utils.FormatDateShort(chat.Gist[0].FirstMessageData),
			),
		})
	}

	return chat.Audio, nil
}

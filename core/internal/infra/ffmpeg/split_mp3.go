package ffmpeg

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
)

// SplitMP3 разбивает аудиофайл на части.
func SplitMP3(fileName string, maxSize int64) ([]model.AudioGist, error) {
	log := slog.With("func", "ffmpeg.SplitMP3")

	absIn, errA := filepath.Abs(fileName)
	if errA != nil {
		return nil, fmt.Errorf("absolute path error for %q: %w", fileName, errA)
	}
	absIn = filepath.ToSlash(absIn)

	// Генерируем шаблон: "input.mp3" -> "input_%03d.mp3"
	dir := filepath.Dir(fileName)
	base := filepath.Base(fileName)
	ext := filepath.Ext(base)
	nameWithoutExt := base[:len(base)-len(ext)]
	outPattern := filepath.Join(dir, fmt.Sprintf("%s_%%03d%s", nameWithoutExt, ext))
	absOutPattern, errB := filepath.Abs(outPattern)
	if errB != nil {
		return nil, fmt.Errorf("absolute path error for %q: %w", outPattern, errB)
	}
	absOutPattern = filepath.ToSlash(absOutPattern)
	log.Debug("output pattern", slog.String("pattern", outPattern))

	// Получаем битрейт для расчета segment_time
	bitrate, errB := getBitrate(absIn)
	if errB != nil {
		return nil, fmt.Errorf("get bitrate error: %w", errB)
	}

	// Примерная длительность для maxSize (bytes/sec = bitrate/8)
	bytesPerSec := float64(bitrate) / 8
	segmentTime := float64(maxSize) / bytesPerSec
	if segmentTime < 60 {
		segmentTime = 60
	}
	log.Debug("calculated segment_time", slog.Float64("s", segmentTime), slog.Int64("bitrate", bitrate))

	success := false
	defer func() {
		if success {
			if errR := os.Remove(fileName); errR != nil {
				log.Error("failed to remove input audio file", slog.String("file", fileName), slog.Any("err", errR))
			}
		}
	}()

	// вызов FFmpeg создаёт все части сразу
	cmd := exec.Command("ffmpeg",
		"-i", absIn,
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%.3f", segmentTime),
		"-c", "copy",
		"-map", "0",
		"-reset_timestamps", "1",
		"-y",
		absOutPattern,
	)

	if errR := cmd.Run(); errR != nil {
		return nil, fmt.Errorf("ffmpeg split error: %w (args: %v)", errR, cmd.Args)
	}

	// Собираем созданные файлы
	result := make([]model.AudioGist, 0)
	outIndex := 0
	for {
		outName := filepath.Join(dir, fmt.Sprintf("%s_%03d%s", nameWithoutExt, outIndex, ext))
		info, err := os.Stat(outName)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			log.Warn("stat failed", slog.String("file", outName), slog.Any("err", err))
			continue
		}

		log.Info("created part",
			slog.Int("index", outIndex),
			slog.Int64("size", info.Size()),
			slog.Int64("max", maxSize))

		result = append(result, model.AudioGist{
			AudioFile: outName,
			Caption:   "",
		})
		outIndex++
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no output files created")
	}

	success = true
	log.Info("split complete", slog.Int("parts", len(result)))
	return result, nil
}

// getBitrate функция получения битрейта из файла
func getBitrate(file string) (int64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-select_streams", "a:0", // первый аудиопоток
		"-show_entries", "stream=bit_rate",
		"-of", "csv=p=0",
		file,
	)
	out, err := cmd.Output()
	if err != nil {
		return 128000, fmt.Errorf("ffprobe bitrate failed: %w", err)
	}

	bitrateStr := strings.TrimSpace(string(out))
	bitrate, err := strconv.ParseInt(bitrateStr, 10, 64)
	if err != nil || bitrate == 0 {
		return 128000, nil // fallback CBR 128kbps
	}

	return bitrate, nil
}

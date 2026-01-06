package ffmpeg

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// ConcatMP3
func ConcatMP3(files []string, out string) error {

	log := slog.With("func", "ffmpeg.ConcatMP3")

	// создаём временный файл со списком
	f, errT := os.CreateTemp("", "ffmpeg-list-*.txt")
	if errT != nil {
		return fmt.Errorf("ffmpeg.ConcatMP3 create temp file error: %w", errT)
	}
	defer func() {
		errR := os.Remove(f.Name())
		if errR != nil {
			log.Error("error removing temp file", errR)
		}
	}()
	defer func() {
		errC := f.Close()
		if errC != nil {
			log.Error("error closing temp file", errC)
		}
	}()

	var b strings.Builder
	for _, name := range files {
		b.WriteString("file '")
		b.WriteString(name)
		b.WriteString("'\n")
	}

	if _, err := f.WriteString(b.String()); err != nil {
		return err
	}

	// ffmpeg -f concat -safe 0 -i list.txt -c copy output.mp3
	cmd := exec.Command("ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", f.Name(),
		"-c", "copy",
		out,
	)

	return cmd.Run()
}

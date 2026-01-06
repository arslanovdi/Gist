package llm_tts

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/arslanovdi/Gist/core/internal/infra/ffmpeg"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"google.golang.org/genai"
)

type Params struct {
	Filename     string `json:"filename,omitempty"` // Имя файла, в который сохраняем аудиопересказ батча
	LanguageCode string `json:"language_code,omitempty"`
	VoiceName    string `json:"voice_name,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
}

const audioDir = "audio" // Для аудиофайлов отдельная папка.

// GenerateAudioGist выполняет запрос к LLM - сценарий GenerateAudioGistFlow для каждого батча.
// Генерирует аудиопересказ чата, по батчам. Сохраняет в mp3 файлы. Имена файлов сохраняются в chat по указателю.
func (s *GenkitService) GenerateAudioGist(ctx context.Context, chat *model.Chat) error {

	log := slog.With("func", "llm_tts.GenkitService.GenerateAudioGist")

	if len(chat.Gist) == 0 {
		return fmt.Errorf("batchGist is empty")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error("get chat gist panic", slog.Any("panic", r))
		}
	}()

	ctxFlow, cancel := context.WithTimeout(ctx, s.flowTimeout) // Общий тайм-аут на обработку всех батчей
	defer cancel()

	for i := range chat.Gist { // генерируем аудиопересказ для каждого батча отдельно
		filename, errF := s.generateAudioGistFlow.Run(ctxFlow,
			Params{
				Filename:     fmt.Sprintf("%s lastmessageID:%d", chat.Title, chat.LastReadMessageID),
				LanguageCode: s.languageCode,
				VoiceName:    s.voiceName,
				Prompt:       chat.Gist[i].Gist,
			})
		if errF != nil {
			return fmt.Errorf("llm-tts.GenerateAudioGist err: %w", errF)
		}

		chat.Gist[i].AudioFile = filename
	}

	return nil
}

// defineGenerateAudioGistFlow  определяет сценарий для генерации краткого пересказа чата.
//
//nolint:gocognit
func (s *GenkitService) defineGenerateAudioGistFlow() {

	log := slog.With("func", "llm_tts.GenkitService.GenerateAudioGistFlow")

	// Определяем сценарий, генерирующий аудио из текста.
	s.generateAudioGistFlow = genkit.DefineFlow(s.g, "GenerateAudioGistFlow", func(ctx context.Context, input Params) (string, error) {
		resp, err := genkit.Generate(ctx, s.g,
			ai.WithConfig(&genai.GenerateContentConfig{
				Temperature:        genai.Ptr[float32](1.0),
				ResponseModalities: []string{"AUDIO"},
				SpeechConfig: &genai.SpeechConfig{
					VoiceConfig: &genai.VoiceConfig{
						PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
							VoiceName: input.VoiceName,
						},
					},
					LanguageCode: input.LanguageCode,
				},
			}),
			ai.WithPrompt(input.Prompt))
		if err != nil {
			return "", err
		}

		// Получаем data URI
		dataURI := resp.Text()
		log.Debug("Received data URI", slog.String("uri", dataURI[:100]+"..."))

		// Парсим data URI и извлекаем PCM данные + sample rate
		pcmData, sampleRate, err := parseDataURI(dataURI)
		if err != nil {
			return "", err
		}

		log.Debug("Parsed PCM data",
			slog.Int("pcm_size", len(pcmData)),
			slog.Int64("sample_rate", int64(sampleRate)),
		)

		// Конвертируем PCM в WAV
		wavData, err := pcmToWAV(pcmData, sampleRate, 1) // 1 = mono
		if err != nil {
			return "", err
		}

		// Сохраняем WAV файл
		wavPath := filepath.Join(audioDir, input.Filename+".wav")
		err = os.WriteFile(wavPath, wavData, 0644)
		if err != nil {
			return "", err
		}

		log.Debug("WAV file saved", slog.String("file", wavPath), slog.Int("size", len(wavData)))

		// Конвертируем WAV в mp3
		mp3path := filepath.Join(audioDir, input.Filename+".mp3")
		errM := ffmpeg.ConvertWavToMp3(wavPath, mp3path)
		if errM != nil {
			return "", errM
		}

		// удаляем wav файл
		errR := os.Remove(wavPath)
		if errR != nil {
			log.Error("error removing temp WAV file", errR)
		}

		return mp3path, nil
	})
}

package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/arslanovdi/Gist/core/internal/adapters/out/llm/tts"
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

// GenerateAudioGist выполняет запрос к LLM - сценарий GenerateAudioGistFlow для каждого батча.
// Генерирует аудиопересказ чата, по батчам. Сохраняет в mp3 файлы. Имена файлов сохраняются в chat по указателю.
func (s *GenkitService) GenerateAudioGist(ctx context.Context, chat *model.Chat) error {

	log := slog.With("func", "llm.GenerateAudioGist")

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

	i := 0
	for i < len(chat.Gist) { // генерируем аудиопересказ для каждого батча отдельно
		log.Debug("start generate audio", slog.Int("last message id", chat.Gist[i].LastMessageID))
		now := time.Now()

		filename, errF := s.generateAudioGistFlow.Run(ctxFlow,
			Params{
				Filename:     fmt.Sprintf("%d_%d", chat.ID, chat.Gist[i].LastMessageID),
				LanguageCode: s.languageCode,
				VoiceName:    s.voiceName,
				Prompt:       chat.Gist[i].Gist,
			})
		if errF != nil {
			if errors.Is(errF, model.ErrResourceExhausted) {

				log.Info("ResourceExhausted, re-init genkit", slog.Int("GeminiApiKeyIndex", s.currentGeminiApiKeyIndex))

				s.initGenkit(context.Background(), true)

				continue
			}
			return fmt.Errorf("llm.GenerateAudioGist err: %w", errF)
		}

		log.Debug("audio generate success",
			slog.Int("last message id", chat.Gist[i].LastMessageID),
			slog.String("filename", filename),
			slog.Any("время генерации", time.Since(now).String()))

		chat.Gist[i].AudioFile = filename
		i++
	}

	return nil
}

// defineGenerateAudioGistFlow  определяет сценарий для генерации краткого пересказа чата.
//
//nolint:gocognit
func (s *GenkitService) defineGenerateAudioGistFlow() {

	log := slog.With("func", "llm_tts.GenkitService.GenerateAudioGistFlow")

	type promptInput struct {
		Text string `json:"text"`
	}

	// Определяем простой запрос(prompt) generateAudioGistPrompt
	generateAudioGistPrompt := genkit.DefinePrompt(s.g, "generateAudioGistPrompt", // TODO можно ли передавать параметры в prompt?
		ai.WithPrompt("{{text}}"),
		ai.WithInputType(promptInput{}),
		ai.WithOutputFormat(ai.OutputFormatText), // выходные данные
		ai.WithConfig(&genai.GenerateContentConfig{
			Temperature:        genai.Ptr[float32](1.0),
			ResponseModalities: []string{"AUDIO"},
			SpeechConfig: &genai.SpeechConfig{
				VoiceConfig: &genai.VoiceConfig{
					PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
						VoiceName: s.voiceName,
					},
				},
				LanguageCode: s.languageCode,
			},
		}),
		ai.WithModelName("googleai/"+s.cfg.LLM.TTS.Gemini.Model), // TODO для TTS жестко задан провайдер Google AI, если появится что-то другое унифицировать эту часть кода.
	)

	// Определяем сценарий, генерирующий аудио из текста.
	s.generateAudioGistFlow = genkit.DefineFlow(s.g, "generateAudioGistFlow", func(ctx context.Context, input Params) (string, error) {
		// выполняем простой запрос с Retry wrapper для обработки 429
		resp, err := retryPrompt(ctx, generateAudioGistPrompt, promptInput{Text: input.Prompt}, log)
		if err != nil {
			return "", fmt.Errorf("generateAudioGistFlow.generateAudioGistPrompt: %w", err)
		}

		// Получаем data URI
		dataURI := resp.Text()
		log.Debug("Received data URI", slog.String("uri", dataURI[:100]+"..."))

		// Парсим data URI и извлекаем PCM данные + sample rate
		pcmData, sampleRate, err := tts.ParseDataURI(dataURI)
		if err != nil {
			return "", err
		}

		log.Debug("Parsed PCM data",
			slog.Int("pcm_size", len(pcmData)),
			slog.Int64("sample_rate", int64(sampleRate)),
		)

		// Конвертируем PCM в WAV
		wavData, err := tts.PcmToWAV(pcmData, sampleRate, 1) // 1 = mono
		if err != nil {
			return "", err
		}

		// Сохраняем WAV файл
		wavPath := filepath.Join(s.cfg.Project.AudioPath, input.Filename+".wav")
		err = os.WriteFile(wavPath, wavData, 0644)
		if err != nil {
			return "", err
		}

		log.Debug("WAV file saved", slog.String("file", wavPath), slog.Int("size", len(wavData)))

		// Конвертируем WAV в mp3
		mp3path := filepath.Join(s.cfg.Project.AudioPath, input.Filename+".mp3")
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

package router

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// TTSHandler воспроизвести аудио-пересказ сути чата.
type TTSHandler struct {
	*BaseHandler
}

// NewTTSHandler конструктор обработчика воспроизведения аудио.
func NewTTSHandler(base *BaseHandler) *TTSHandler {
	return &TTSHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *TTSHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionTTS
}

// Handle Реализация интерфейса CallbackHandler
func (h *TTSHandler) Handle(ctx *th.Context, query telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "TTSHandler")
	log.Debug("handling TTS callback")

	// Обязательно сразу отвечаем, что обработчик работает, могут быть проблемы из-за медленных ответов > 10 секунд
	_ = h.Bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))

	audioGist, errA := h.CoreService.GetAudioGist(ctx, payload.ChatID, payload.Page) // получаем имя файла с нужным аудиопересказом
	if errA != nil {
		return fmt.Errorf("tgbot.router.TTSHandler Ошибка получения аудиофайла: %w", errA)
	}

	for i := range audioGist { // Отправляем файлы, по очереди
		audioFile, errO := os.Open(audioGist[i].AudioFile)
		if errO != nil {
			return fmt.Errorf("open audio file error: %w", errO)
		}
		defer func() {
			errC := audioFile.Close()
			if errC != nil {
				log.Error("audio file close error:", errC)
			}
		}()

		// Отправляем аудио
		_, errS := h.Bot.SendVoice(ctx, tu.Voice( // Отправка голосового сообщения.
			tu.ID(h.UserID),
			tu.File(audioFile),
		).WithCaption(audioGist[i].Caption))
		if errS != nil {
			return errS
		}
	}

	/*_, errS := h.Bot.SendAudio(ctx, tu.Audio(	// отправка mp3 файла
		tu.ID(h.UserID),
		tu.File(audioFile),
	))
	if errS != nil {
		return errS
	}*/

	return nil
}

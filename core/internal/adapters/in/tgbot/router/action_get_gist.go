package router

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// GistHandler получает краткий пересказ чата.
type GistHandler struct {
	*BaseHandler
}

// NewGistHandler конструктор обработчика кнопки генерации краткого пересказа чата.
func NewGistHandler(base *BaseHandler) *GistHandler {
	return &GistHandler{BaseHandler: base}
}

// CanHandle Реализация интерфейса CallbackHandler
func (h *GistHandler) CanHandle(payload *CallbackPayload) bool {
	return payload.Action == ActionGetGist
}

// Handle Реализация интерфейса CallbackHandler
func (h *GistHandler) Handle(ctx *th.Context, query telego.CallbackQuery, payload *CallbackPayload) error {
	log := slog.With("func", "GistHandler")
	log.Debug("handling get gist callback")

	processing := func(message string, part int, llm bool) { // callback функция для оповещения о прогрессе выполнения.
		if llm {
			bar := strings.Repeat("█", part/10) + strings.Repeat("░", 10-part/10)
			_ = h.editMessage(ctx,
				fmt.Sprintf("%s\n\n [%s] %d%%", message, bar, part))
		} else {
			_ = h.editMessage(ctx,
				fmt.Sprintf("%s\n\n %d сообщений загружено", message, part))
		}
	}

	// Обязательно сразу отвечаем, что обработчик работает, могут быть проблемы из-за медленных ответов > 10 секунд
	_ = h.Bot.AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID)) //.WithText("⏳ Генерируем пересказ..."))

	_, errG := h.CoreService.GetChatGist(ctx, payload.ChatID, processing) // Получаем краткий пересказ, сохраняем его в кэш.
	if errG != nil {
		log.Error("GetChatGist", slog.Any("error", errG))
	}

	chatDetail, errD := h.CoreService.GetChatDetail(ctx, payload.ChatID) // Получаем информацию о чате из кэша
	if errD != nil {
		chatDetail = &model.Chat{}
		log.Error("GetChatDetail", slog.Any("error", errD))
	}

	return h.showChatDetail(ctx, chatDetail, payload.Src, 1) // gistPage после генерации = 1

}

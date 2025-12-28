// Package router реализует меню телеграм бота
package router

import (
	"context"
	"fmt"
	"log/slog"
	"unsafe"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// CoreService определяет интерфейс для взаимодействия с бизнес-логикой.
type CoreService interface {
	GetAllChats(ctx context.Context) ([]model.Chat, error)                         // Возвращает список всех чатов пользователя.
	GetChatsWithUnreadMessages(ctx context.Context) ([]model.Chat, error)          // Возвращает список чатов с непрочитанными сообщениями.
	GetFavoriteChats(ctx context.Context) ([]model.Chat, error)                    // Возвращает список избранных чатов.
	GetChatGist(ctx context.Context, chatID int64) ([]model.BatchGist, error)      // Возвращает короткий пересказ непрочитанных сообщений чата.
	GetChatDetail(ctx context.Context, chatID int64) (*model.Chat, error)          // Получение информации о чате из кэша
	ChangeFavorites(ctx context.Context, chatID int64) error                       // Добавление чата в избранное
	MarkAsRead(ctx context.Context, chatID int64, pageID int) (*model.Chat, error) // Отмечает указанный чат как прочитанный, удаляя из кэша прочитанный пересказ. Возвращает обновленный объект чата.
}

// CallbackHandler определяет интерфейс для обработчиков колбэков от инлайн кнопок
type CallbackHandler interface {
	CanHandle(payload *CallbackPayload) bool
	Handle(ctx *th.Context, query telego.CallbackQuery, payload *CallbackPayload) error
}

// CallbackRouter обработчик колбэков (роутер), который уже вызывает нужный метод. Управляет маршрутизацией колбэков
type CallbackRouter struct {
	log *slog.Logger

	handlers []CallbackHandler
}

// NewCallbackRouter создает новый экземпляр CallbackRouter
func NewCallbackRouter() *CallbackRouter {
	return &CallbackRouter{
		handlers: make([]CallbackHandler, 0),
		log:      slog.Default().With("component", "callback_router"),
	}
}

// RegisterHandler регистрирует обработчик колбэков
func (r *CallbackRouter) RegisterHandler(handler CallbackHandler) {
	r.handlers = append(r.handlers, handler)
}

// Handle маршрутизация колбэков.
func (r *CallbackRouter) Handle(ctx *th.Context, query telego.CallbackQuery) error {
	if query.Data == "" {
		return fmt.Errorf("no callback data")
	}

	payload, err := parseCallback(query.Data)
	if err != nil {
		r.log.Error("failed to parse callback data", "error", err, "data", query.Data)
		return fmt.Errorf("invalid callback data: %w", err)
	}

	fmt.Println(payload) // TODO delete this
	fmt.Println(unsafe.Sizeof(payload))

	for _, handler := range r.handlers { // Перебор зарегистрированных колбэков
		if handler.CanHandle(payload) {
			return handler.Handle(ctx, query, payload)
		}
	}

	r.log.Warn("no handler found for callback", "payload", payload)
	return nil
}

// ShowMainMenu выводит главное меню в Telegram боте.
func (r *CallbackRouter) ShowMainMenu(ctx *th.Context) error {
	// Создаем фейковый callback с MenuMain
	payload, err := CallbackPayload{Menu: MenuMain}.String()
	if err != nil {
		return fmt.Errorf("show main menu failed: %w", err)
	}

	return r.Handle(ctx, telego.CallbackQuery{Data: payload})
}

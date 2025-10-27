package tgbot

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// authorizedOnlyMiddleware middleware для проверки прав доступа пользователя.
func (b *Bot) authorizedOnlyMiddleware() th.Handler {
	return func(ctx *th.Context, update telego.Update) (err error) {
		var from telego.User
		log := slog.With("func", "tgbot.authorizedOnlyMiddleware")

		// Определяем отправителя из возможных источников
		switch {
		case update.Message != nil:
			from = *update.Message.From
		case update.CallbackQuery != nil:
			from = update.CallbackQuery.From
		case update.InlineQuery != nil:
			from = update.InlineQuery.From
		default:
			// Неизвестный или неподдерживаемый тип обновления
			return nil
		}

		// Если пользователь не определен или ID не совпадает - игнорируем запрос
		if from.ID != b.allowedUserID {
			log.Debug("Unauthorized access attempt", slog.Int64("user_id", from.ID))
			return nil
		}

		// Разрешаем обработку
		return ctx.Next(update)
	}
}

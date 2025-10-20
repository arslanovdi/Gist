package tgbot

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (b *Bot) Authentication(ctx context.Context, userID int64, phone, code chan<- string, authError chan error) error {
	log := slog.With("func", "bot.Authentication", "user_id", userID)

	b.userCredentials[userID] = &model.Credential{}

	// Этап 1: Запрос номера телефона
	log.Debug("Starting authentication process")
	b.userStates[userID] = model.AuthGetPhone // Инициализируем состояние пользователя

	if err := b.askForPhone(ctx, userID); err != nil {
		return fmt.Errorf("failed to request phone: %w", err)
	}
	if err := b.waitForPhoneInput(ctx, userID); err != nil {
		return fmt.Errorf("failed to get phone: %w", err)
	}

	phone <- b.userCredentials[userID].Phone
	log.Debug("Authentication, phone sent successfully")

	// Этап 3: Запрос кода подтверждения
	if err := b.askForCode(ctx, userID); err != nil {
		return fmt.Errorf("failed to request code: %w", err)
	}
	if err := b.waitForCodeInput(ctx, userID); err != nil {
		return fmt.Errorf("failed to get code: %w", err)
	}

	code <- b.userCredentials[userID].Code
	log.Debug("Authentication, code sent successfully")

	// Удаляем пользователя из мап.
	delete(b.userStates, userID)
	delete(b.userCredentials, userID)

	// Ожидаем подтверждение успешной авторизации
	AuthError := <-authError
	if AuthError != nil {
		log.Debug("Authentication failure")
		_, err := b.bot.SendMessage(ctx, tu.Messagef(
			tu.ID(userID),
			fmt.Sprintf("Authentication Error %s", AuthError),
		))
		if err != nil {
			log.Error("Failed to send message", slog.Any("error", err))
		}
		return AuthError
	}

	log.Debug("Authentication completed successfully")
	_, err := b.bot.SendMessage(ctx, tu.Messagef(
		tu.ID(userID),
		"🎉 Авторизация успешно завершена!\n\nТеперь вы можете использовать все функции бота.",
	))
	if err != nil {
		log.Error("Failed to send message", slog.Any("error", err))
	}

	return nil
}

// waitForPhoneInput ожидает ввод номера телефона
func (b *Bot) waitForPhoneInput(ctx context.Context, userID int64) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)

			// Проверяем, изменилось ли состояние пользователя
			state, exists := b.userStates[userID]
			if !exists {
				return fmt.Errorf("user state not found")
			}
			if state == model.AuthGetCode {
				return nil
			}
		}
	}
}

// waitForCodeInput ожидает ввод кода подтверждения
func (b *Bot) waitForCodeInput(ctx context.Context, userID int64) error {
	for { // TODO продумать как этот цикл завершится если пользователь не завершив аутентификацию введет команду /start.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)

			// Проверяем, изменилось ли состояние пользователя
			state, exists := b.userStates[userID]
			if !exists {
				return fmt.Errorf("user state not found")
			}
			if state == model.AuthDone {
				return nil
			}
		}
	}
}

// AuthMessageHandler обрабатывает сообщения во время авторизации
func (b *Bot) AuthMessageHandler(ctx *th.Context, message telego.Message) error {
	userID := message.Chat.ID
	userState, exists := b.userStates[userID]
	if !exists {
		return nil // Игнорируем сообщения не в процессе авторизации
	}

	log := slog.With("func", "bot.AuthMessageHandler", "user_id", userID, "state", userState)

	switch userState {
	case model.AuthGetPhone:
		return b.handlePhoneInput(ctx, message)
	case model.AuthGetCode:
		return b.handleCodeInput(ctx, message)
	default:
		log.Error("Unknown user state", slog.Any("state", userState))
		return nil
	}
}

// handlePhoneInput обрабатывает ввод номера телефона
func (b *Bot) handlePhoneInput(ctx *th.Context, message telego.Message) error {
	userID := message.Chat.ID
	var phone string

	// Проверяем, отправил ли пользователь контакт
	if message.Contact != nil && message.Contact.PhoneNumber != "" {
		phone = message.Contact.PhoneNumber
	} else if message.Text != "" {
		phone = strings.TrimSpace(message.Text)
	} else {
		return b.sendValidationError(ctx, userID, "Пожалуйста, введите номер телефона или используйте кнопку \"Отправить номер телефона\"")
	}

	// Валидация номера телефона
	_, err := b.validatePhone(phone)
	if err != nil {
		return b.sendValidationError(ctx, userID, fmt.Sprintf("❌ %s\n\nПожалуйста, введите номер в правильном формате:\nПример: +1234567890", err.Error()))
	}

	// Сохраняем номер телефона и переходим к следующему этапу
	log := slog.With("user_id", userID, "phone", phone)
	log.Debug("Phone number received and validated")

	cred, ok := b.userCredentials[userID]
	if !ok {
		log.Error("User credentials not found")
		return b.sendValidationError(ctx, userID, fmt.Sprintf("❌ Внутренняя ошибка сервера. \n\n Пожалуйста, попробуйте еще раз с помощью команды /start"))
	}
	cred.Phone = phone
	b.userStates[userID] = model.AuthGetCode

	// Отправляем подтверждение
	_, errS := b.bot.SendMessage(ctx, tu.Messagef(
		tu.ID(userID),
		"✅ Номер телефона принят",
	))
	if errS != nil {
		log.Error("Failed to send message", slog.Any("error", errS))
	}

	return err
}

// handleCodeInput обрабатывает ввод кода подтверждения
func (b *Bot) handleCodeInput(ctx *th.Context, message telego.Message) error {
	userID := message.Chat.ID
	code := strings.TrimSpace(message.Text)

	if code == "" {
		return b.sendValidationError(ctx, userID, "Пожалуйста, введите код подтверждения")
	}

	// Валидация кода
	if err := b.validateCode(code); err != nil {
		return b.sendValidationError(ctx, userID, fmt.Sprintf("❌ %s\n\nПожалуйста, введите код из сообщения Telegram", err.Error()))
	}

	log := slog.With("user_id", userID, "code_length", len(code))
	log.Debug("Verification code received and validated")

	cred, ok := b.userCredentials[userID]
	if !ok {
		log.Error("User credentials not found")
		return b.sendValidationError(ctx, userID, fmt.Sprintf("❌ Внутренняя ошибка сервера. \n\n Пожалуйста, попробуйте еще раз с помощью команды /start"))
	}
	cred.Code = code
	b.userStates[userID] = model.AuthDone

	return nil
}

// sendValidationError отправляет сообщение об ошибке валидации
func (b *Bot) sendValidationError(ctx *th.Context, userID int64, errorMsg string) error {
	_, err := b.bot.SendMessage(ctx, tu.Messagef(
		tu.ID(userID),
		"%s",
		errorMsg,
	))
	return err
}

// askForPhone запрашивает номер телефона у пользователя
func (b *Bot) askForPhone(ctx context.Context, userID int64) error {
	_, err := b.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: userID},
		Text: "🚀 Добро пожаловать! Для начала работы мне нужно получить доступ к вашему Telegram аккаунту.\n\n" +
			"Пожалуйста, введите ваш номер телефона в международном формате:\n" +
			"📞 Пример: +1234567890",
		ReplyMarkup: &telego.ReplyKeyboardMarkup{
			Keyboard: [][]telego.KeyboardButton{
				{
					{
						Text:           "📱 Отправить номер телефона",
						RequestContact: true,
					},
				},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
		},
	})
	return err
}

// askForCode запрашивает код подтверждения
func (b *Bot) askForCode(ctx context.Context, userID int64) error {

	_, err := b.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: userID},
		Text: fmt.Sprintf("📱 Код подтверждения отправлен на номер \n\n" +
			"Пожалуйста, введите 5-значный код из сообщения:"), // TODO в сообщении 5-значный код, тогда как в валидации 5-7 знаков. Сколько все таки знаков в коде подтверждения.
		ReplyMarkup: &telego.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
	return err
}

// TODO подумать над переходом к валидации при помощи пакета "github.com/nyaruka/phonenumbers"
func (b *Bot) validatePhone(phone string) (string, error) {

	// Приводим к единому формату без разделителей
	reClean := regexp.MustCompile(`[\s()+-]`)
	phone = reClean.ReplaceAllString(phone, "")

	// Проверяем российские форматы номеров
	re := regexp.MustCompile(`^[78]?(\d{10})$`)
	if !re.MatchString(phone) {
		return "", fmt.Errorf("неверный формат номера телефона")
	}
	return phone, nil // TODO Нужно ли возвращать номер без + ?
}

func (b *Bot) validateCode(code string) error {

	codePattern := regexp.MustCompile(`^\d{5,7}$`)

	// Коды Telegram обычно содержат 5-7 цифр
	if !codePattern.MatchString(code) {
		return fmt.Errorf("код должен состоять от 5 до 7 цифр")
	}
	return nil
}

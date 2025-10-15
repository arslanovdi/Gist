package user

import (
	"github.com/arslanovdi/Gist/core/internal/infra/config"
	"github.com/arslanovdi/Gist/core/internal/tgbot"
	"github.com/gotd/td/telegram"
)

type User struct {
	appID   int
	appHash string

	client *telegram.Client // Это клиентское подключение.
}

var _ tgbot.UserCommands = (*User)(nil)

func (u *User) GetAllChats() ([]string, error) {
	// TODO implement me
	panic("implement me")
}

func (u *User) GetGistFromChat(chatID int) (string, error) {
	// TODO implement me
	panic("implement me")
}

func New(cfg *config.Config) *User {
	/*client := telegram.NewClient(cfg.Bot.AppID, cfg.Bot.AppHash, telegram.Options{
		PublicKeys:          nil,
		DC:                  0,
		DCList:              dcs.List{},
		Resolver:            nil,
		NoUpdates:           false,
		ReconnectionBackoff: nil,
		OnDead:              nil,
		MigrationTimeout:    0,
		Random:              nil,
		Logger:              nil,
		SessionStorage:      nil,
		UpdateHandler:       nil,
		Middlewares:         nil,
		AckBatchSize:        0,
		AckInterval:         0,
		RetryInterval:       0,
		MaxRetries:          0,
		ExchangeTimeout:     0,
		DialTimeout:         0,
		CompressThreshold:   0,
		Device:              telegram.DeviceConfig{},
		MessageID:           nil,
		Clock:               nil,
		TracerProvider:      nil,
		OnTransfer:          nil,
		OnSelfError:         nil,
	})*/

	return &User{
		appID:   cfg.Bot.AppID,
		appHash: cfg.Bot.AppHash,
		// client:  client,
	}
}

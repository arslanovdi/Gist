package tgbot

import (
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func (b *Bot) AnyMessage(_ *th.Context, message telego.Message) error {
	fmt.Println("Message:", message.Text)
	return nil
}

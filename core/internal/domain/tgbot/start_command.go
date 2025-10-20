package tgbot

import (
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (b *Bot) StartCommand(ctx *th.Context, update telego.Update) error {

	state, exist := b.userStates[update.Message.Chat.ID] // Запущен процесс аутентификации TODO нужно прерывать запущенную аутентификацию...?
	if exist {
		return fmt.Errorf("user already in state %s", state)
	}

	chats, errG := b.coreService.GetAllChats(ctx, update.Message.Chat.ID)
	if errG != nil {
		return errG
	}

	fmt.Println(chats)

	_, errS := b.bot.SendMessage(ctx, tu.Messagef(
		tu.ID(update.Message.Chat.ID),
		"Hello %s!", update.Message.From.FirstName,
	))
	return errS
}

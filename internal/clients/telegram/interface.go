package telegram

import "github.com/PaulSonOfLars/gotgbot/v2"

type Client interface {
	HandleUpdate(u *gotgbot.Update) error
	Auth(authUrl string) error
}

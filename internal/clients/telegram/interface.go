package telegram

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Client interface {
	HandleUpdate(ctx context.Context, u *gotgbot.Update) error
	Auth(ctx context.Context, code string, state string) error
}

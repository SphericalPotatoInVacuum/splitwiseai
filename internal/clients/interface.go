package clients

import (
	"splitwiseai/internal/clients/mindee"
	"splitwiseai/internal/clients/telegram"
)

type Clients interface {
	Mindee() mindee.Client
	Telegram() telegram.Client
}

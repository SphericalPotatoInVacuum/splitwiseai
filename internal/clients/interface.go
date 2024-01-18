package clients

import (
	"splitwiseai/internal/clients/openai"
	"splitwiseai/internal/clients/telegram"
)

type Clients interface {
	Telegram() telegram.Client
	OpenAI() openai.Client
}

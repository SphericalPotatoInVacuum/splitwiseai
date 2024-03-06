package clients

import (
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/mq/tgupdatesmq"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/openai"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/telegram"
)

type Clients interface {
	Telegram() telegram.Client
	OpenAI() openai.Client
	TgUpdatesMQ() tgupdatesmq.Client
}

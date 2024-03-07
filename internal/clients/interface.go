package clients

import (
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/mq/tgupdatesmq"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/ocr"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/openai"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/splitwise"
)

type Clients interface {
	OpenAI() openai.Client
	TgUpdatesMQ() tgupdatesmq.Client
	Splitwise() splitwise.Client
	OCR() ocr.Client
}

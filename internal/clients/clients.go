package clients

import (
	"fmt"

	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/mq/tgupdatesmq"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/ocr"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/openai"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/splitwise"
)

type Config struct {
	MindeeCfg            ocr.Config
	SplitwiseCfg         splitwise.Config
	OpenAICfg            openai.Config
	TelegramUpdatesMQCfg tgupdatesmq.Config

	OcrClient string `env:"OCR_CLIENT"`
}

type clients struct {
	oai         openai.Client
	ocr         ocr.Client
	splitwise   splitwise.Client
	tgUpdatesMQ tgupdatesmq.Client
}

func NewClients(cfg Config) (Clients, error) {
	splitwiseClient, err := splitwise.NewClient(cfg.SplitwiseCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create splitwise client: %w", err)
	}

	oaiClient, err := openai.NewClient(cfg.OpenAICfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create openai client: %w", err)
	}

	var mindeeClient ocr.Client
	if cfg.OcrClient == "mindee" {
		mindeeClient, err = ocr.NewClient(cfg.MindeeCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create mindee client: %w", err)
		}
	} else if cfg.OcrClient == "gpt" {
		mindeeClient = oaiClient
	} else if cfg.MindeeCfg.Enabled {
		return nil, fmt.Errorf("invalid OCR client: %s", cfg.OcrClient)
	}

	tgUpdatesMQClient, err := tgupdatesmq.NewClient(cfg.TelegramUpdatesMQCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tgupdatesmq client: %w", err)
	}

	return &clients{
		oai:         oaiClient,
		ocr:         mindeeClient,
		splitwise:   splitwiseClient,
		tgUpdatesMQ: tgUpdatesMQClient,
	}, nil
}

func (c *clients) OpenAI() openai.Client {
	return c.oai
}

func (c *clients) OCR() ocr.Client {
	return c.ocr
}

func (c *clients) Splitwise() splitwise.Client {
	return c.splitwise
}

func (c *clients) TgUpdatesMQ() tgupdatesmq.Client {
	return c.tgUpdatesMQ
}

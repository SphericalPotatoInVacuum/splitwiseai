package clients

import (
	"fmt"

	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/db"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/mq/tgupdatesmq"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/ocr"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/openai"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/splitwise"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/telegram"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/models"
)

type Config struct {
	MindeeCfg            ocr.Config
	SplitwiseCfg         splitwise.Config
	DBCfg                db.Config
	ModelsCfg            models.Config
	TelegramCfg          telegram.Config
	OpenAICfg            openai.Config
	TelegramUpdatesMQCfg tgupdatesmq.Config

	OcrClient string `env:"OCR_CLIENT"`
}

type clients struct {
	oai         openai.Client
	telegram    telegram.Client
	tgUpdatesMQ tgupdatesmq.Client
}

func NewClients(cfg Config) (Clients, error) {
	splitwiseClient, err := splitwise.NewClient(cfg.SplitwiseCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create splitwise client: %w", err)
	}

	db := db.NewClient(cfg.DBCfg)

	ms, err := models.NewModels(db, cfg.ModelsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create models: %w", err)
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
	} else {
		return nil, fmt.Errorf("invalid OCR client: %s", cfg.OcrClient)
	}

	telegramClient, err := telegram.NewClient(cfg.TelegramCfg, &telegram.BotDeps{
		Models:    ms,
		Splitwise: splitwiseClient,
		Ocr:       mindeeClient,
		OpenAI:    oaiClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram client: %w", err)
	}

	tgUpdatesMQClient, err := tgupdatesmq.NewClient(cfg.TelegramUpdatesMQCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tgupdatesmq client: %w", err)
	}

	return &clients{
		telegram:    telegramClient,
		oai:         oaiClient,
		tgUpdatesMQ: tgUpdatesMQClient,
	}, nil
}

func (c *clients) Telegram() telegram.Client {
	return c.telegram
}

func (c *clients) OpenAI() openai.Client {
	return c.oai
}

func (c *clients) TgUpdatesMQ() tgupdatesmq.Client {
	return c.tgUpdatesMQ
}

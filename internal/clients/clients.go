package clients

import (
	"fmt"

	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/db/tokensdb"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/db/usersdb"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/ocr"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/openai"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/splitwise"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/telegram"
)

type Config struct {
	MindeeCfg    ocr.Config
	SplitwiseCfg splitwise.Config
	UsersDbCfg   usersdb.Config
	TokensDbCfg  tokensdb.Config
	TelegramCfg  telegram.Config
	OpenAICfg    openai.Config

	OcrClient string `env:"OCR_CLIENT"`
}

type clients struct {
	oai      openai.Client
	telegram telegram.Client
}

func NewClients(cfg Config) (Clients, error) {
	splitwiseClient, err := splitwise.NewClient(cfg.SplitwiseCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create splitwise client: %w", err)
	}

	usersDbClient, err := usersdb.NewClient(cfg.UsersDbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create usersdb client: %w", err)
	}

	tokensDbClient, err := tokensdb.NewClient(cfg.TokensDbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokensdb client: %w", err)
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
		UsersDb:   usersDbClient,
		TokensDb:  tokensDbClient,
		Splitwise: splitwiseClient,
		Ocr:       mindeeClient,
		OpenAI:    oaiClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram client: %w", err)
	}

	return &clients{
		telegram: telegramClient,
		oai:      oaiClient,
	}, nil
}

func (c *clients) Telegram() telegram.Client {
	return c.telegram
}

func (c *clients) OpenAI() openai.Client {
	return c.oai
}

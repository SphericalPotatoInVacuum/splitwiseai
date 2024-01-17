package clients

import (
	"fmt"
	"splitwiseai/internal/clients/db/tokensdb"
	"splitwiseai/internal/clients/db/usersdb"
	"splitwiseai/internal/clients/mindee"
	"splitwiseai/internal/clients/splitwise"
	"splitwiseai/internal/clients/telegram"
)

type Config struct {
	MindeeCfg    mindee.Config
	SplitwiseCfg splitwise.Config
	UsersDbCfg   usersdb.Config
	TokensDbCfg  tokensdb.Config
	TelegramCfg  telegram.Config
}

type clients struct {
	mindee   mindee.Client
	usersDb  usersdb.Client
	tokensDb tokensdb.Client
	telegram telegram.Client
}

func NewClients(cfg Config) (Clients, error) {
	splitwiseClient, err := splitwise.NewClient(cfg.SplitwiseCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create splitwise client: %w", err)
	}

	mindeeClient, err := mindee.NewClient(cfg.MindeeCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create mindee client: %w", err)
	}

	usersDbClient, err := usersdb.NewClient(cfg.UsersDbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create usersdb client: %w", err)
	}

	tokensDbClient, err := tokensdb.NewClient(cfg.TokensDbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokensdb client: %w", err)
	}

	telegramClient, err := telegram.NewClient(cfg.TelegramCfg, &telegram.BotDeps{
		UsersDb:   usersDbClient,
		TokensDb:  tokensDbClient,
		Splitwise: splitwiseClient,
		Mindee:    mindeeClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram client: %w", err)
	}

	return &clients{
		mindee:   mindeeClient,
		usersDb:  usersDbClient,
		telegram: telegramClient,
	}, nil
}

func (c *clients) Mindee() mindee.Client {
	return c.mindee
}

func (c *clients) Telegram() telegram.Client {
	return c.telegram
}

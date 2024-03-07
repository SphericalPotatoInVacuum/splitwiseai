package main

import (
	"context"
	"fmt"

	"github.com/SphericalPotatoInVacuum/splitwiseai/handlers/ext"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients"
	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

type TelegramUpdateHandlerDeps struct {
	Clients clients.Clients
}

var telegramUpdateHandlerDeps *TelegramUpdateHandlerDeps = nil

type TelegramUpdateHandlerConfig struct {
	ClientsCfg clients.Config
}

func InitTelegramUpdateHandler() *TelegramUpdateHandlerDeps {
	ext.Init()

	if telegramUpdateHandlerDeps != nil {
		return telegramUpdateHandlerDeps
	}

	zap.S().Debug("Initialising function dependencies")

	cfg := TelegramUpdateHandlerConfig{}
	if err := env.Parse(&cfg); err != nil {
		zap.S().Panicw("Failed to load the config from the the environment", zap.Error(err))
	}

	cs, err := clients.NewClients(cfg.ClientsCfg)
	if err != nil {
		zap.S().Panicw("Failed to create clients", zap.Error(err))
	}

	if err != nil {
		zap.S().Panicw("Failed to create bot client", zap.Error(err))
	}

	return &TelegramUpdateHandlerDeps{
		Clients: cs,
	}
}

func HandleTelegramUpdate(ctx context.Context, event ext.APIGatewayRequest) (*ext.APIGatewayResponse, error) {
	deps := InitTelegramUpdateHandler()
	var err error

	zap.S().Infow("Handling telegram update", "event", event, "context", ctx)

	err = deps.Clients.TgUpdatesMQ().PublishMessage(ctx, event.Body, map[string]string{"type": "telegram_update"})
	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %v", err)
	}

	return &ext.APIGatewayResponse{
		StatusCode:      200,
		Headers:         map[string]string{"content-type": "application/json"},
		Body:            "",
		IsBase64Encoded: false,
	}, nil
}

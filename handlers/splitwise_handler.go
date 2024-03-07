package main

import (
	"context"
	"encoding/json"

	"github.com/SphericalPotatoInVacuum/splitwiseai/handlers/ext"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients"
	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

type SplitwiseHandlerDeps struct {
	Clients clients.Clients
}

var splitwiseHandlerDeps *SplitwiseHandlerDeps = nil

type SplitwiseHandlerConfig struct {
	ClientsCfg clients.Config
}

func InitSplitwiseHandler() *SplitwiseHandlerDeps {
	ext.Init()

	if splitwiseHandlerDeps != nil {
		return splitwiseHandlerDeps
	}

	zap.S().Debug("Initialising function dependencies")

	cfg := SplitwiseHandlerConfig{}
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

	return &SplitwiseHandlerDeps{
		Clients: cs,
	}
}

func HandleSplitwiseCallback(ctx context.Context, event ext.APIGatewayRequest) (*ext.APIGatewayResponse, error) {
	deps := InitSplitwiseHandler()

	code, ok := event.QueryStringParameters["code"]
	if !ok {
		return &ext.APIGatewayResponse{
			StatusCode:      403,
			Headers:         map[string]string{"content-type": "application/json"},
			Body:            "{\"error\": \"code query param is missing\"}",
			IsBase64Encoded: false,
		}, nil
	}

	state, ok := event.QueryStringParameters["state"]
	if !ok {
		return &ext.APIGatewayResponse{
			StatusCode:      403,
			Headers:         map[string]string{"content-type": "application/json"},
			Body:            "{\"error\": \"state query param is missing\"}",
			IsBase64Encoded: false,
		}, nil
	}

	messageBytes, err := json.Marshal(ext.SplitwiseCallback{Code: code, State: state})
	if err != nil {
		zap.S().Panicw("Failed to marshal message", zap.Error(err))
	}

	err = deps.Clients.TgUpdatesMQ().PublishMessage(ctx, string(messageBytes), map[string]string{"type": "splitwise_callback"})

	if err != nil {
		zap.S().Errorw("Failed to publish message", zap.Error(err))
		return &ext.APIGatewayResponse{
			StatusCode:      502,
			Headers:         map[string]string{"content-type": "application/json"},
			Body:            "{\"error\": \"failed to process callback\"}",
			IsBase64Encoded: false,
		}, nil
	}

	return &ext.APIGatewayResponse{
		StatusCode:      200,
		Headers:         map[string]string{"content-type": "text/html; charset=utf-8"},
		Body:            "<h1>Данные отправлены! Можете закрыть эту страницу.</h1>",
		IsBase64Encoded: false,
	}, nil
}

package main

import (
	"context"
	"fmt"

	"github.com/SphericalPotatoInVacuum/splitwiseai/handlers/ext"
	"go.uber.org/zap"
)

func HandleSplitwiseCallback(ctx context.Context, event APIGatewayRequest) (*APIGatewayResponse, error) {
	deps := ext.Init()

	code, ok := event.QueryStringParameters["code"]
	if !ok {
		return &APIGatewayResponse{
			StatusCode:      403,
			Headers:         map[string]string{"content-type": "application/json"},
			Body:            "{\"error\": \"code query param is missing\"}",
			IsBase64Encoded: false,
		}, nil
	}

	state, ok := event.QueryStringParameters["state"]
	if !ok {
		return &APIGatewayResponse{
			StatusCode:      403,
			Headers:         map[string]string{"content-type": "application/json"},
			Body:            "{\"error\": \"state query param is missing\"}",
			IsBase64Encoded: false,
		}, nil
	}

	err := deps.Clients.Telegram().Auth(ctx, code, state)
	if err != nil {
		zap.S().Errorw("Error handling splitwise callback", zap.Error(err))
		return &APIGatewayResponse{
			StatusCode:      502,
			Headers:         map[string]string{"content-type": "application/json"},
			Body:            "{\"error\": \"failed to authenticate with Telegram\"}",
			IsBase64Encoded: false,
		}, nil
	}
	return &APIGatewayResponse{
		StatusCode:      200,
		Headers:         map[string]string{"content-type": "text/html; charset=utf-8"},
		Body:            "<h1>Данные отправлены! Можете закрыть эту страницу.</h1>",
		IsBase64Encoded: false,
	}, nil
}

func HandleTelegramUpdate(ctx context.Context, event APIGatewayRequest) (*APIGatewayResponse, error) {
	deps := ext.Init()
	var err error

	zap.S().Infow("Handling telegram update", "event", event, "context", ctx)

	err = deps.Clients.TgUpdatesMQ().PublishMessage(ctx, event.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %v", err)
	}

	return &APIGatewayResponse{
		StatusCode:      200,
		Headers:         map[string]string{"content-type": "application/json"},
		Body:            "",
		IsBase64Encoded: false,
	}, nil
}

package main

import (
	"context"
	"encoding/json"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/SphericalPotatoInVacuum/splitwiseai/handlers/ext"
	"go.uber.org/zap"
)

func HandleTelegramUpdateMessage(ctx context.Context, event Event) (string, error) {
	deps := ext.Init()

	for _, message := range event.Messages {
		update := &gotgbot.Update{}
		err := json.Unmarshal([]byte(message.Details.Message.Body), update)
		if err != nil {
			zap.S().Errorw("Failed to unmarshal message", "message", message, zap.Error(err))
			return "Failed to unmarshal message", err
		}

		err = deps.Clients.Telegram().HandleUpdate(ctx, update)
		if err != nil {
			zap.S().Errorw("Failed to handle update", "update", update, zap.Error(err))
			return "Failed to handle update", err
		}
	}

	return "OK", nil
}

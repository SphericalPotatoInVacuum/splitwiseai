package main

import (
	"context"
	"encoding/json"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/SphericalPotatoInVacuum/splitwiseai/handlers/ext"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/bot"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/db"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/models"
	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

type TgProcessorDeps struct {
	Bot bot.Client
}

var tgProcessorDeps *TgProcessorDeps = nil

type TgProcessorConfig struct {
	ClientsCfg clients.Config
	ModelsCfg  models.Config
	DBCfg      db.Config
	BotCfg     bot.Config
}

func InitTgProcessor() *TgProcessorDeps {
	ext.Init()

	if tgProcessorDeps != nil {
		return tgProcessorDeps
	}

	zap.S().Debug("Initialising function dependencies")

	cfg := TgProcessorConfig{}
	if err := env.Parse(&cfg); err != nil {
		zap.S().Panicw("Failed to load the config from the the environment", zap.Error(err))
	}

	cs, err := clients.NewClients(cfg.ClientsCfg)
	if err != nil {
		zap.S().Panicw("Failed to create clients", zap.Error(err))
	}

	db := db.NewClient(cfg.DBCfg)

	models, err := models.NewModels(db, cfg.ModelsCfg)
	if err != nil {
		zap.S().Panicw("Failed to create models", zap.Error(err))
	}

	bot, err := bot.NewClient(cfg.BotCfg, &bot.BotDeps{
		Clients: cs,
		Models:  models,
	})
	if err != nil {
		zap.S().Panicw("Failed to create bot client", zap.Error(err))
	}

	return &TgProcessorDeps{
		Bot: bot,
	}
}

func ProcessTelegramUpdate(ctx context.Context, event ext.Event) (string, error) {
	deps := InitTgProcessor()

	for _, message := range event.Messages {
		if message.Details.Message.MessageAttributes["type"].StringValue == "telegram_update" {
			update := &gotgbot.Update{}
			err := json.Unmarshal([]byte(message.Details.Message.Body), update)
			if err != nil {
				zap.S().Errorw("Failed to unmarshal message", "message", message, zap.Error(err))
				return "Failed to unmarshal message", err
			}

			err = deps.Bot.HandleUpdate(ctx, update)
			if err != nil {
				zap.S().Errorw("Failed to handle update", "update", update, zap.Error(err))
				return "Failed to handle update", err
			}
		} else if message.Details.Message.MessageAttributes["type"].StringValue == "splitwise_callback" {
			splitwiseCallback := ext.SplitwiseCallback{}
			err := json.Unmarshal([]byte(message.Details.Message.Body), &splitwiseCallback)
			if err != nil {
				zap.S().Errorw("Failed to unmarshal message", "message", message, zap.Error(err))
				return "Failed to unmarshal message", err
			}
			err = deps.Bot.Auth(ctx, splitwiseCallback.Code, splitwiseCallback.State)
			if err != nil {
				zap.S().Errorw("Failed to auth", "splitwiseCallback", splitwiseCallback, zap.Error(err))
				return "Failed to auth", err
			}
		}
	}

	return "OK", nil
}

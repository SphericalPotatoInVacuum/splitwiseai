package ext

import (
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/logging"
	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

type FunctionDeps struct {
	Clients clients.Clients
}

var initialised = false
var deps *FunctionDeps = nil

func Init() *FunctionDeps {
	var err error

	if initialised {
		return deps
	}

	zap.ReplaceGlobals(logging.CreateLogger())

	zap.S().Debug("Initialising function dependencies")

	zap.S().Debug("Loading config from the environment")
	cfg := clients.Config{}
	if err := env.Parse(&cfg); err != nil {
		zap.S().Panicw("Failed to load the config from the the environment", zap.Error(err))
	}

	zap.S().Debug("Creating clients")
	cs, err := clients.NewClients(cfg)
	if err != nil {
		zap.S().Panicw("Failed to create clients", zap.Error(err))
	}

	initialised = true
	deps = &FunctionDeps{
		Clients: cs,
	}

	zap.S().Debug("Initialised function dependencies")

	return deps
}

package ext

import (
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/logging"
	"go.uber.org/zap"
)

var initialised = false

func Init() {
	if initialised {
		return
	}

	zap.ReplaceGlobals(logging.CreateLogger())

	initialised = true
}

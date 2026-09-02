package repository

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"demo_bot/internal/app/repository/state"
)

var Module = fx.Module(
	"repository",
	fx.Options(
		state.Module,
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("repository")
	}),
)

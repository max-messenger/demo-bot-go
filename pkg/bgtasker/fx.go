package bgtasker

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"bg_tasker",
	fx.Provide(
		NewPool,
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("bg_tasker")
	}),
	fx.Invoke(func(lc fx.Lifecycle, pool *Pool) {
		lc.Append(fx.Hook{
			OnStart: pool.Start,
			OnStop:  pool.Stop,
		})
	}),
)

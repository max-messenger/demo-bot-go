package grace

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"grace",
	fx.Provide(
		NewServicePool,
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("grace")
	}),
)

type Params struct {
	fx.In

	Logger   *zap.Logger
	Services []Service `group:"grace"`
}

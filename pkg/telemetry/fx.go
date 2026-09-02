package telemetry

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"telemetry",
	fx.Provide(
		NewTelemetry,
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("telemetry")
	}),

	fx.Invoke(func(lc fx.Lifecycle, srv *Telemetry) {
		lc.Append(fx.Hook{
			OnStop: srv.Stop,
		})
	}),
)

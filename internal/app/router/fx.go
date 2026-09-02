package router

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"httpsrv",
	fx.Provide(
		NewRouter,
		ControllersAdapter,
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("http_server")
	}),
)

type (
	ControllersIn struct {
		fx.In

		Controllers []Controller `group:"controller"`
	}

	ControllersOut struct {
		fx.Out

		Controllers Controllers
	}
)

func ControllersAdapter(in ControllersIn) ControllersOut {
	return ControllersOut{
		Controllers: in.Controllers,
	}
}

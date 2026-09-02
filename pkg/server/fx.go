package server

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewServersProvider,
		NewHTTPServer,
		NewGRPCServer,
		NewSystemServer,
	),
	fx.Invoke(func(lc fx.Lifecycle, srv *Servers) {
		lc.Append(fx.Hook{
			OnStart: srv.Start,
			OnStop:  srv.Stop,
		})
	}),
)

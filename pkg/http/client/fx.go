package client

import "go.uber.org/fx"

var Module = fx.Module(
	"httpclient",
	fx.Provide(
		New,
	),
)

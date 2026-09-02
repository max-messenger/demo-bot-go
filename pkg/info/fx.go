package info

import "go.uber.org/fx"

var Module = fx.Module(
	"info",
	fx.Provide(
		NewInfo,
	),
)

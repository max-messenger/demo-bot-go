package state

import (
	"go.uber.org/fx"

	repostate "demo_bot/internal/app/repository/state"
)

var Module = fx.Module(
	"state_service",
	fx.Provide(
		Adapter,
	),
)

type (
	AdapterIn struct {
		fx.In

		Repo *repostate.Repository
	}

	AdapterOut struct {
		fx.Out

		Service Service
	}
)

func Adapter(in AdapterIn) AdapterOut {
	return AdapterOut{
		Service: NewService(in.Repo),
	}
}

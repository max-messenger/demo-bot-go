package bot

import (
	"fmt"

	api "github.com/max-messenger/maxbot"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"demo_bot/internal/app/services/bot/scenario"
	"demo_bot/internal/app/services/state"
	"demo_bot/pkg/connections/maxbot"
	"demo_bot/pkg/grace"
	"demo_bot/pkg/locker"
)

var Module = fx.Module(
	"bot",
	fx.Provide(
		New,
		Adapter,
		apiAdapter,
		scenarioAPIAdapter,
		lockerAdapter,
		graceAdapter,
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("bot")
	}),
)

type graceOut struct {
	fx.Out

	Service grace.Service `group:"grace"`
}

func graceAdapter(sub *MaxBot) graceOut {
	return graceOut{
		Service: sub,
	}
}

type (
	AdapterIn struct {
		fx.In

		Client *api.Api
		Store  state.Service
	}

	AdapterOut struct {
		fx.Out

		Client Client
		Store  Store
	}
)

func Adapter(in AdapterIn) AdapterOut {
	return AdapterOut{
		Client: in.Client,
		Store:  in.Store,
	}
}

func apiAdapter(pool *maxbot.Pool) (*api.Api, error) {
	a, err := pool.GetPool("public")
	if err != nil {
		return nil, fmt.Errorf("get maxbot api pool: %w", err)
	}

	return a, nil
}

func scenarioAPIAdapter(a *api.Api) scenario.API {
	return a
}

func lockerAdapter(l *locker.Locker) Locker {
	return l
}

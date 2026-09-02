//nolint:ireturn
package maxbot

import (
	"github.com/max-messenger/maxbot"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"demo_bot/internal/app/router"
	"demo_bot/pkg/http/client"
)

var Module = fx.Module(
	"max_bot_pool",
	fx.Provide(
		NewMaxBotPool,
		adapter,
		AdapterControllerOut,
		AdapterOptOut,
		configProvider("max_bot"),
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("max_bot_pool")
	}),
)

func configProvider(name string) func(cp config.Provider) (PoolConfig, error) {
	return func(cp config.Provider) (PoolConfig, error) {
		cfg := PoolConfig{}

		if err := cp.Get(name).Populate(&cfg); err != nil {
			return PoolConfig{}, err
		}

		return cfg, nil
	}
}

func adapter(log *zap.Logger) HTTPClient {
	opts := []client.Option{
		client.WithLogger(log.Named("maxbot")),
	}

	return client.New("maxbot", opts...)
}

type PoolParams struct {
	fx.In

	Configs     PoolConfig
	Opts        []maxbot.Opt            `group:"max_opts"`
	Middlewares []maxbot.MiddlewareFunc `group:"max_middlewares"`
}

type OptOut struct {
	fx.Out

	Options maxbot.Opt `group:"max_opts"`
}

func AdapterOptOut(cli HTTPClient) OptOut {
	return OptOut{
		Options: maxbot.WithHTTPClient(cli),
	}
}

type ControllerOut struct {
	fx.Out

	Controller router.Controller `group:"controller"`
}

func AdapterControllerOut(ctrl *Pool) ControllerOut {
	return ControllerOut{
		Controller: ctrl,
	}
}

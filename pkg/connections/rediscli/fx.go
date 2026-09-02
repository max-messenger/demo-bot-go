package rediscli

import (
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"rediscli",
	fx.Provide(
		NewPool,
		configProvider("redis"),
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("rediscli")
	}),
	fx.Invoke(
		func(lc fx.Lifecycle, pp *Pool) {
			lc.Append(fx.Hook{
				OnStop: pp.Stop,
			})
		},
	),
)

func configProvider(name string) func(cp config.Provider) (PoolConfig, error) {
	return func(cp config.Provider) (PoolConfig, error) {
		cfg := PoolConfig{}

		if err := cp.Get(name).Populate(&cfg); err != nil {
			return PoolConfig{}, err
		}

		for connName, connCfg := range cfg {
			cfg[connName] = connCfg.Prepare()
		}

		return cfg, nil
	}

}

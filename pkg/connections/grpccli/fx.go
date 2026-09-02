package grpccli

import (
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"grpc-conn",
	fx.Provide(
		NewPool,
		configProvider("grpc_clients"),
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("grpc-conn")
	}),
	fx.Invoke(func(lc fx.Lifecycle, pool *Pool) {
		lc.Append(fx.Hook{
			OnStop: pool.Stop,
		})
	}),
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

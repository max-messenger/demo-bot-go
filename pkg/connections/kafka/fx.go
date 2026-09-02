package kafka

import (
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	Module = fx.Module(
		"kafka",
		fx.Provide(
			NewPool,
			configProvider("kafka"),
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("kafka")
		}),
	)
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

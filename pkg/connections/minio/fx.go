package minio

import (
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"minio",
	fx.Provide(
		NewClient,
		configProvider("minio"),
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("minio")
	}),
)

func configProvider(name string) func(cp config.Provider) (*Config, error) {
	return func(cp config.Provider) (*Config, error) {
		cfg := &Config{}

		if err := cp.Get(name).Populate(cfg); err != nil {
			return nil, err
		}

		return cfg, nil
	}
}

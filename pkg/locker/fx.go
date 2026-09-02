package locker

import (
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"locker",
	fx.Provide(
		New,
		configProvider("locker"),
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("locker")
	}),
)

func configProvider(name string) func(cp config.Provider) (Config, error) {
	return func(cp config.Provider) (Config, error) {
		cfg := NewConfig()

		if err := cp.Get(name).Populate(&cfg); err != nil {
			return Config{}, err
		}

		return cfg, nil
	}

}

package bigbro

import (
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"demo_bot/pkg/http/client"
)

var Module = fx.Module(
	"bigbro",
	fx.Provide(
		New,
		adapter,
		configProvider("bigbro"),
	),
	fx.Decorate(func(log *zap.Logger) *zap.Logger {
		return log.Named("bigbro")
	}),
)

func adapter(log *zap.Logger) HTTPClient {
	opts := []client.Option{
		client.WithLogger(log.Named("bigbro")),
	}

	return client.New("bigbro", opts...)
}

func configProvider(name string) func(cp config.Provider) (Config, error) {
	return func(cp config.Provider) (Config, error) {
		cfg := Config{}

		if err := cp.Get(name).Populate(&cfg); err != nil {
			return Config{}, err
		}

		return cfg, nil
	}
}

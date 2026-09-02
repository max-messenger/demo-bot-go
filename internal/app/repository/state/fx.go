package state

import (
	"fmt"
	"time"

	"go.uber.org/config"
	"go.uber.org/fx"

	"demo_bot/pkg/connections/rediscli"
)

// Config holds state repository configuration.
type Config struct {
	TTL time.Duration `yaml:"ttl"`
}

var Module = fx.Module(
	"state",
	fx.Provide(
		NewRepository,
		Adapter,
		configProvider("state"),
	),
)

type (
	AdapterIn struct {
		fx.In

		Store *rediscli.Pool
	}

	AdapterOut struct {
		fx.Out

		Store Store
	}
)

func Adapter(in AdapterIn) (AdapterOut, error) {
	client, err := in.Store.GetPool("main")
	if err != nil {
		return AdapterOut{}, err
	}

	return AdapterOut{
		Store: client,
	}, nil
}

func configProvider(name string) func(cp config.Provider) (Config, error) {
	return func(cp config.Provider) (Config, error) {
		cfg := Config{}

		if err := cp.Get(name).Populate(&cfg); err != nil {
			return Config{}, fmt.Errorf("populate state config: %w", err)
		}

		return cfg, nil
	}
}

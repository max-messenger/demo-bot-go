package ratelimiter

import (
	"go.uber.org/config"
	"go.uber.org/fx"

	"demo_bot/pkg/connections/rediscli"
)

var Module = fx.Module(
	"rate_limiter",
	fx.Provide(
		adapterLimitGet,
		configProvider("rate_limiter"),
	),
)

func configProvider(name string) func(cp config.Provider) (Config, error) {
	return func(cp config.Provider) (Config, error) {
		cfg := Config{}

		if err := cp.Get(name).Populate(&cfg); err != nil {
			return Config{}, err
		}

		return cfg, nil
	}

}

type adapterLimitGetIn struct {
	fx.In

	LimitGet []CustomLimitFunc `group:"rate_limit_custom"`
}

func adapterLimitGet(in adapterLimitGetIn) CustomLimitFuncs {
	return in.LimitGet
}

var ModuleLocal = fx.Module(
	"local_rate_limiter",
	fx.Provide(
		NewLocalLimiter,
	),
)

var ModuleShared = fx.Module(
	"shared_rate_limiter",
	fx.Provide(
		NewSharedLimiter,
		sharedRateLimiterRedisAdapter,
	),
)

func sharedRateLimiterRedisAdapter(cfg Config, rediscliPool *rediscli.Pool) (RedisClient, error) {
	sharedCfg := cfg.Shared

	return rediscliPool.GetPool(sharedCfg.PoolName)
}

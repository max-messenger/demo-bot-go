package locker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"demo_bot/pkg/connections/rediscli"
)

const (
	lockPrefixKey = "locker"
)

var (
	unlockScript = redis.NewScript(`
		local key = KEYS[1]
		local value = ARGV[1]

		if redis.call("get", key) == value then
			return redis.call("del", key)
		end

		return 0
	`)
)

var (
	ErrMaxRetries = errors.New("max retries")
)

type Key struct {
	Component string
	Key       string
}

type Locker struct {
	cfg Config

	rd *rediscli.Redis
}

func New(config Config, pool *rediscli.Pool) (*Locker, error) {
	rd, err := pool.GetPool(config.PoolName)
	if err != nil {
		return nil, fmt.Errorf("get redis pool: %w", err)
	}

	return &Locker{
		rd:  rd,
		cfg: config,
	}, nil
}

func (l *Locker) Lock(
	ctx context.Context,
	key Key,
	value string,
	ttl time.Duration,
) (func(ctx context.Context) error, error) {
	sKey := serializeKey(key)
	err := l.rd.Call(ctx, "lock", func(
		ctx context.Context,
		ucl redis.UniversalClient,
		keyFormatter rediscli.KeyFormatter,
	) error {
		keyFormatted := keyFormatter.FormatKey(sKey)
		for retries := l.cfg.Retries; retries > 0; retries-- {
			set, qerr := ucl.SetNX(ctx, keyFormatted, value, l.ttl(ttl)).Result()
			if qerr != nil {
				return fmt.Errorf("set nx: %w", qerr)
			}

			if set {
				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(l.cfg.Delay):
			}

		}

		return ErrMaxRetries
	})

	if err != nil {
		return nil, fmt.Errorf("try lock: %w", err)
	}

	return l.releaseFunc(sKey, value), nil
}

func (l *Locker) releaseFunc(key string, value string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := l.rd.Call(ctx, "unlock", func(
			ctx context.Context,
			ucl redis.UniversalClient,
			keyFormatter rediscli.KeyFormatter,
		) error {
			keyFormatted := keyFormatter.FormatKey(key)
			_, qErr := unlockScript.Eval(ctx, ucl, []string{keyFormatted}, value).Result()
			if qErr != nil {
				if errors.Is(qErr, redis.Nil) {
					return nil
				}

				return fmt.Errorf("unlock: %w", qErr)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("unlock: %w", err)
		}

		return nil
	}
}

func (l *Locker) ttl(dur time.Duration) time.Duration {
	if dur <= 0 {
		return l.cfg.TTL
	}

	return dur
}

func serializeKey(key Key) string {
	return fmt.Sprintf("%s:%s:%s", lockPrefixKey, key.Component, key.Key)
}

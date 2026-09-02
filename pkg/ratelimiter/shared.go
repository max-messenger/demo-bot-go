package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"demo_bot/pkg/connections/rediscli"
	"demo_bot/pkg/telemetry"
)

var (
	tokenBucketScript = redis.NewScript(`
		local key = KEYS[1]
		local rate = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local requested = tonumber(ARGV[4])

		local info = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(info[1])
		local last_refill = tonumber(info[2])

		if tokens == nil then
			tokens = capacity
			last_refill = now
		end

		local delta = math.max(0, now - last_refill)
		local filled_tokens = math.min(capacity, tokens + delta * rate)
		local allowed = filled_tokens >= requested and 1 or 0

		if allowed == 1 then
			filled_tokens = filled_tokens - requested
		end

		redis.call('HMSET', key, 'tokens', filled_tokens, 'last_refill', now)
		redis.call('EXPIRE', key, math.ceil(capacity / rate) + 1)

		return allowed
	`)
)

type RedisClient interface {
	Call(ctx context.Context, name string, callback rediscli.CallCallback) error
}

type SharedLimiter struct {
	config Config

	lm limiter

	redis  RedisClient
	logger *zap.Logger
}

func NewSharedLimiter(
	config Config,
	logger *zap.Logger,
	redis *rediscli.Redis,
	customLimits CustomLimitFuncs,
) *SharedLimiter {
	return &SharedLimiter{
		config: config,
		lm:     newLimiter(config.Rate, customLimits),
		logger: logger,
		redis:  redis,
	}
}

func (l *SharedLimiter) Limit(ctx context.Context, key Key, action Action) (err error) {
	defer func(t time.Time) {
		status := telemetry.ErrLabelValue(err)
		if errors.Is(err, ErrLimitExceeded) {
			status = "true"
		}

		sharedTotal.WithLabelValues(
			string(action),
			status,
		).Inc()

		if err != nil && !errors.Is(err, ErrLimitExceeded) {
			sharedFails.WithLabelValues(string(action)).Inc()
		}

		sharedDuration.WithLabelValues(string(action)).Observe(time.Since(t).Seconds())
	}(time.Now())

	limit := l.lm.LimitForKeyAndAction(key, action)
	if limit == LimitUnlimited {
		return nil
	}

	// Convert limit to rate per second and burst capacity
	ratePerSec := float64(limit)
	capacity := int64(limit)

	var allowed int64
	err = l.redis.Call(
		ctx,
		"token_bucket",
		func(ctx context.Context, cli redis.UniversalClient, kf rediscli.KeyFormatter) error {
			redisKey := kf.FormatKey(fmt.Sprintf("%s:ratelimit:%s:%s", l.config.Shared.Prefix, action, key))

			// Get current time in seconds with millisecond precision for smooth rate limiting
			now := float64(time.Now().UnixNano()) / float64(time.Second)

			// Execute Lua script with token bucket logic
			result, qerr := tokenBucketScript.Run(ctx, cli, []string{redisKey},
				ratePerSec, // rate: tokens per second
				capacity,   // capacity: maximum tokens
				now,        // current time
				1,          // requested: tokens to consume
			).Result()

			if qerr != nil {
				return fmt.Errorf("failed to execute token bucket script: %w", qerr)
			}

			var ok bool
			allowed, ok = result.(int64)
			if !ok {
				return fmt.Errorf("failed to parse script result")
			}

			return nil
		})

	if err != nil {
		return fmt.Errorf("call: %w", err)
	}

	if allowed == 0 {
		sharedExceed.WithLabelValues(string(action)).Inc()

		return ErrLimitExceeded
	}

	return nil
}

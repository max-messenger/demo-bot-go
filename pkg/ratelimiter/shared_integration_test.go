package ratelimiter_test

import (
	"context"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"demo_bot/pkg/connections/rediscli"
	"demo_bot/pkg/ratelimiter"
)

var (
	defaultRedisSeeds = "127.0.0.1:7001,127.0.0.1:7002,127.0.0.1:7003,127.0.0.1:7004,127.0.0.1:7005,127.0.0.1:7006"
	redisSeeds        = os.Getenv("REDIS_SEEDS")
)

func TestIntegration_Shared(t *testing.T) {
	t.Skip()
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))
	ctx := context.Background()

	seeds := redisSeeds
	if seeds == "" {
		seeds = defaultRedisSeeds
	}
	cfg := &rediscli.Config{
		KeyPrefix: "test",
		Addrs:     strings.Split(seeds, ","),
	}

	rds, err := rediscli.NewRedis(logger, cfg, "test")
	require.NoError(t, err)
	require.NotNil(t, rds)

	testDuration := 3 * time.Second

	tests := []struct {
		Name     string
		Limit    ratelimiter.Limit
		Duration time.Duration
	}{
		{Name: "11/1", Limit: 11, Duration: testDuration},
		{Name: "211/1", Limit: 211, Duration: testDuration},
		{Name: "1000/1", Limit: 1000, Duration: testDuration},
	}
	t.Run("cases", func(t *testing.T) {
		for i := range tests {
			test := tests[i]

			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				sw := ratelimiter.NewSharedLimiter(
					ratelimiter.Config{
						Rate: map[string]ratelimiter.RateConfig{
							"test": {
								Keys: map[string]float64{
									test.Name: float64(test.Limit),
								},
							},
						},
					},
					logger,

					rds,
					nil,
				)
				processed := 0
				timer := time.NewTimer(test.Duration)

			forever:
				for {
					select {
					case <-timer.C:
						break forever
					default:
						if err := sw.Limit(ctx, ratelimiter.Key(test.Name), "test"); err == nil {
							processed++
						} else if !errors.Is(err, ratelimiter.ErrLimitExceeded) {
							require.Nil(t, err)
						}
					}
				}

				actual := float64(processed) / float64(test.Duration.Seconds())
				expected := float64(test.Limit)

				t.Log(test.Name, processed, actual, expected, math.Round(math.Abs(1-actual/expected)))

				const epsilon = 0.1

				require.LessOrEqual(t, math.Round(math.Abs(1-actual/expected)), epsilon)

			})
		}
	})
}

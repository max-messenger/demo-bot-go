package ratelimiter_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"demo_bot/pkg/ratelimiter"
)

func TestUnit_LocalLimiter_Table(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))
	ctx := context.Background()

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

				sw := ratelimiter.NewLocalLimiter(
					ratelimiter.Config{
						Local: ratelimiter.LocalConfig{
							TTL: time.Second,
						},
						Rate: map[string]ratelimiter.RateConfig{
							"test": {
								Limit: int(test.Limit),
							},
						},
					},
					logger,
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

func TestUnit_LocalLimiter_Custom(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))
	ctx := context.Background()

	testDuration := 3 * time.Second
	limit := 100000

	ll := ratelimiter.NewLocalLimiter(
		ratelimiter.Config{
			Local: ratelimiter.LocalConfig{
				TTL: time.Second,
				Custom: map[string]ratelimiter.LocalCustomConfig{
					"custom": {
						CacheSize: 2,
					},
				}},
		},
		logger,

		ratelimiter.CustomLimitFuncs{
			func() (ratelimiter.Action, ratelimiter.LimitGet) {
				return ratelimiter.Action("custom"),
					func(
						key ratelimiter.Key,
						action ratelimiter.Action,
					) ratelimiter.Limit {
						return ratelimiter.Limit(limit)
					}
			},
		},
	)

	processed := 0
	timer := time.NewTimer(testDuration)

forever:
	for {
		select {
		case <-timer.C:
			break forever
		default:
			if err := ll.Limit(ctx, "test", "custom"); err == nil {
				processed++
			} else if !errors.Is(err, ratelimiter.ErrLimitExceeded) {
				require.Nil(t, err)
			}
		}
	}

	actual := float64(processed) / float64(testDuration.Seconds())
	expected := float64(limit)

	t.Log(processed, actual, expected, math.Round(math.Abs(1-actual/expected)))

	const epsilon = 0.1

	require.LessOrEqual(t, math.Round(math.Abs(1-actual/expected)), epsilon)
}

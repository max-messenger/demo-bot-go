package rediscli_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"demo_bot/pkg/connections/rediscli"
)

var (
	defaultRedisCluster = "127.0.0.1:7001,127.0.0.1:7002,127.0.0.1:7003,127.0.0.1:7004,127.0.0.1:7005,127.0.0.1:7006"
	redisClusterSeeds   = os.Getenv("REDIS_CLUSTER_SEEDS")

	defaultRedisSingle = "127.0.0.1:6379"
	redisSeed          = os.Getenv("REDIS_SEED")
)

func TestIntegration_Redis_Single(t *testing.T) {
	t.Skip()
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	ctx := context.Background()

	seeds := redisSeed
	if seeds == "" {
		seeds = defaultRedisSingle
	}
	cfg := &rediscli.Config{
		KeyPrefix: "test",
		Addrs:     strings.Split(seeds, ","),
	}

	rds, err := rediscli.NewRedis(logger, cfg, "test")
	require.NoError(t, err)
	require.NotNil(t, rds)

	err = rds.Call(ctx, "test", func(ctx context.Context, ucl redis.UniversalClient, keyFormatter rediscli.KeyFormatter) error {
		key := uuid.NewString()
		return ucl.SetEx(ctx, key, 1, time.Second).Err()
	})

	require.NoError(t, err)
}

func TestIntegration_Redis_Cluster(t *testing.T) {
	t.Skip()
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	ctx := context.Background()

	seeds := redisClusterSeeds
	if seeds == "" {
		seeds = defaultRedisCluster
	}
	cfg := &rediscli.Config{
		KeyPrefix: "test",
		Addrs:     strings.Split(seeds, ","),
	}

	rds, err := rediscli.NewRedis(logger, cfg, "test")
	require.NoError(t, err)
	require.NotNil(t, rds)

	err = rds.Call(ctx, "test", func(ctx context.Context, ucl redis.UniversalClient, keyFormatter rediscli.KeyFormatter) error {
		key := uuid.NewString()
		return ucl.SetEx(ctx, key, 1, time.Second).Err()
	})

	require.NoError(t, err)
}

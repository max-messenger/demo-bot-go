package rediscli

import (
	"context"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"demo_bot/pkg/telemetry"
)

type (
	metricsHook struct {
		name string
	}
)

func newMetricsHook(name string) *metricsHook {
	return &metricsHook{
		name: name,
	}
}

func (h *metricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		poolConnCreatedTotal.WithLabelValues(h.name, addr).Inc()

		return next(ctx, network, addr)
	}
}

func (h *metricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		defer func(t time.Time) {
			singleCommands.WithLabelValues(
				h.name,
				cmd.Name(),
				telemetry.ErrLabelValue(cmd.Err()),
			).Observe(time.Since(t).Seconds())
		}(time.Now())

		err := next(ctx, cmd)
		if err != nil {
			return err
		}

		return nil
	}
}

func (h *metricsHook) ProcessPipelineHook(
	next redis.ProcessPipelineHook,
) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) (err error) {
		defer func(t time.Time) {
			singleCommands.WithLabelValues(
				h.name,
				"pipeline",
				telemetry.ErrLabelValue(err),
			).Observe(time.Since(t).Seconds())

			pipelinedCommands.WithLabelValues(
				h.name,
				telemetry.ErrLabelValue(err),
			).Inc()
		}(time.Now())

		err = next(ctx, cmds)
		if err != nil {
			return err
		}

		return nil
	}
}

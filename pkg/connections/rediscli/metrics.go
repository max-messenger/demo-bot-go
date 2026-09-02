package rediscli

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"demo_bot/pkg/telemetry"
)

const labelName = "name"

var (
	buckets = []float64{.001, .003, .005, .01, .025, .05, .1, .25, .5, 1}

	duration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_pool_action_duration_seconds",
			Help:    "redis pool action request latencies in seconds.",
			Buckets: buckets,
		},
		[]string{labelName, "action", telemetry.ErrLabel},
	)

	connections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_pool_connections",
			Help: "redis pool connections",
		},
		[]string{labelName, "status"},
	)

	connectionsCall = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_pool_connections_calls",
			Help: "Number of redis pool connections calls.",
		},
		[]string{labelName, "status"},
	)

	poolConnCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_pool_connection_created_total",
			Help: "Total number of created connections in pool.",
		},
		[]string{labelName, "addr"},
	)

	singleCommands = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "redis_single_commands",
		Help:    "Histogram of single Redis commands",
		Buckets: telemetry.DefaultHistogramBuckets,
	}, []string{labelName, "command", telemetry.ErrLabel})

	pipelinedCommands = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "redis_pipelined_commands_total",
		Help: "Number of pipelined Redis commands",
	}, []string{labelName, telemetry.ErrLabel})
)

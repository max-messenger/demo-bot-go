package bigbro

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"demo_bot/pkg/telemetry"
)

var (
	metricRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bigbro_requests_total",
			Help: "Total number of BigBro requests.",
		},
		[]string{"method", telemetry.ErrLabel},
	)

	metricRequestsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bigbro_requests_duration_seconds",
			Help:    "BigBro requests latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

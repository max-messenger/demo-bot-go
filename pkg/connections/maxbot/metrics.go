package maxbot

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subscriber_requests_total",
			Help: "Total number of subscriber requests.",
		},
		[]string{"method"},
	)

	metricRequestsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "subscriber_requests_duration_seconds",
			Help:    "subscriber requests latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

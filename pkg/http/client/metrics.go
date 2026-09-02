package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"demo_bot/pkg/telemetry"
)

var (
	// metricHTTPRequestsTotal — общее количество HTTP-запросов по методам и статусам.
	metricHTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_client_requests_total",
			Help: "Total number of HTTP requests made by the client.",
		},
		[]string{"name", "method", "status_code", telemetry.ErrLabel},
	)

	// metricHTTPRequestsDuration — время выполнения HTTP-запросов.
	metricHTTPRequestsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_client_requests_duration_seconds",
			Help:    "HTTP requests latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"name", "method"},
	)
)

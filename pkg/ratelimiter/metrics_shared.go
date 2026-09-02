package ratelimiter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"demo_bot/pkg/telemetry"
)

const (
	labelSharedAction = "action"
	labelSharedStatus = "status"
)

var (
	sharedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shared_rate_limiter_actions_total",
			Help: "Total number of shared rate limiter action requests.",
		},
		[]string{labelSharedAction, labelSharedStatus},
	)

	sharedDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "shared_rate_limiter_action_duration_seconds",
			Help:    "shared rate limiter action request latencies in seconds.",
			Buckets: telemetry.DefaultHistogramBuckets,
		},
		[]string{labelSharedAction},
	)

	sharedExceed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shared_rate_limiter_exceed_total",
			Help: "Total number of shared rate limiter action exceed.",
		},
		[]string{labelSharedAction},
	)

	sharedFails = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shared_rate_limiter_fails_total",
			Help: "Total number of shared rate limiter action fails.",
		},
		[]string{labelSharedAction},
	)
)

package ratelimiter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const labelAction = "action"

var (
	localTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "local_rate_limiter_actions_total",
			Help: "Total number of local rate limiter action requests.",
		},
		[]string{labelAction},
	)

	localExceed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "local_rate_limiter_exceed_total",
			Help: "Total number of local rate limiter action exceed.",
		},
		[]string{labelAction},
	)
)

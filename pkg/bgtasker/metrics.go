package bgtasker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"demo_bot/pkg/telemetry"
)

var (
	queueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bg_tasker_queue_size",
			Help: "Queue size",
		},
		[]string{"name"},
	)

	taskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bg_tasker_task_duration_seconds",
			Help:    "task duration in seconds.",
			Buckets: telemetry.DefaultHistogramBuckets,
		},
		[]string{"action"},
	)
)

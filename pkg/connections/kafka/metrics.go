package kafka

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"demo_bot/pkg/telemetry"
)

const (
	namespaceKafkaPool = "kafka_pool"
	namespaceKafka     = "kafka"
	subsystemGroup     = "consumer_group"
	metricDurationSec  = "duration_seconds"
	labelName          = "name"
	labelGroup         = "group"
	labelTopic         = "topic"
)

var (
	producerTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespaceKafkaPool,
			Subsystem: "producer",
			Name:      "send_count_total",
			Help:      "Produced events count",
		},
		[]string{labelName, labelTopic, telemetry.ErrLabel},
	)

	producerAsyncHandle = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceKafkaPool,
			Subsystem: "producer_async",
			Name:      metricDurationSec,
			Help:      "Time elapsed to async produce single messsage",
			Buckets:   telemetry.DefaultHistogramBuckets,
		},
		[]string{labelName, labelTopic, telemetry.ErrLabel},
	)

	producerAsyncMessagesInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespaceKafkaPool,
			Subsystem: "producer_async",
			Name:      "messages_in_flight",
			Help:      "Number of produced messages in flight",
		},
		[]string{labelName, labelTopic},
	)

	consumerHandle = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceKafkaPool,
			Subsystem: "consumer",
			Name:      metricDurationSec,
			Help:      "Time elapsed to consume single messsage",
			Buckets:   telemetry.DefaultHistogramBuckets,
		},
		[]string{labelName, labelTopic, telemetry.ErrLabel},
	)

	consumerGroupHandle = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceKafkaPool,
			Subsystem: subsystemGroup,
			Name:      metricDurationSec,
			Help:      "Time elapsed to consume single messsage",
			Buckets:   telemetry.DefaultHistogramBuckets,
		},
		[]string{labelName, labelGroup, labelTopic, telemetry.ErrLabel},
	)

	consumerGroupBatchHandle = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceKafkaPool,
			Subsystem: subsystemGroup,
			Name:      "batch_duration_seconds",
			Help:      "Time elapsed to consume batch messsages",
			Buckets:   telemetry.DefaultHistogramBuckets,
		},
		[]string{labelName, labelGroup, labelTopic, telemetry.ErrLabel},
	)

	consumerGroupPollTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespaceKafka,
			Subsystem: subsystemGroup,
			Name:      "poll_total",
			Help:      "Amount of poll received per partition",
		},
		[]string{labelName, labelGroup, labelTopic, "partition", "is_full"},
	)
	consumerGroupBatchPollTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespaceKafka,
			Subsystem: subsystemGroup,
			Name:      "batch_poll_total",
			Help:      "Amount of poll received per partition",
		},
		[]string{labelName, labelGroup, labelTopic, "partition", "is_full"},
	)
)

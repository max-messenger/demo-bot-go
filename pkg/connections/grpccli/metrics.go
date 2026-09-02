package grpccli

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"

	"demo_bot/pkg/telemetry"
)

var (
	clMetrics = grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets(telemetry.DefaultHistogramBuckets),
		),
	)
)

func init() {
	prometheus.MustRegister(clMetrics)
}

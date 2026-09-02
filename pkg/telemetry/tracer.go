package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Telemetry struct {
	cfg    Config
	logger *zap.Logger

	tracerStopFunc func(ctx context.Context) error
}

func NewTelemetry(cfg Config, logger *zap.Logger) (*Telemetry, error) {
	t := &Telemetry{
		cfg:    cfg,
		logger: logger,
	}

	if cfg.CollectorConfig.Enabled {
		stopFunc, err := initTracer(cfg)
		if err != nil {
			return nil, fmt.Errorf("init tracer: %w", err)
		}

		t.tracerStopFunc = stopFunc
	}

	return t, nil
}

func (t *Telemetry) Stop(ctx context.Context) error {
	if t.tracerStopFunc != nil {
		err := t.tracerStopFunc(ctx)
		if err != nil {
			return fmt.Errorf("stop func: %w", err)
		}
	}

	return nil
}

func initTracer(config Config) (func(ctx context.Context) error, error) {
	endpoint := config.CollectorConfig.Endpoint
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var exporter *otlptrace.Exporter
	if config.CollectorConfig.EndpointType == collectorEndpointTypeGRPC {
		conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to endpoint: %w", err)
		}

		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter: %w", err)
		}
	}
	if config.CollectorConfig.EndpointType == collectorEndpointTypeHTTP {
		var err error
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithHeaders(config.CollectorConfig.Headers),
			otlptracehttp.WithURLPath(config.CollectorConfig.URLPath),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter: %w", err)
		}
	}

	if exporter == nil {
		return nil, fmt.Errorf("unknown endpoint type: %s", config.CollectorConfig.EndpointType)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(config.CollectorConfig.SamplingRatio)),
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.Version),
			semconv.ServiceInstanceID(config.Hostname),
			semconv.HostName(config.Hostname),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

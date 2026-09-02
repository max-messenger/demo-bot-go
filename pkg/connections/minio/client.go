package minio

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func NewClient(cfg *Config) (*minio.Client, error) {
	defaultTransport, err := minio.DefaultTransport(cfg.UseSSL)
	if err != nil {
		return nil, fmt.Errorf("create default transport: %w", err)
	}
	raw, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, cfg.SecretToken),
		Secure: cfg.UseSSL,
		Transport: otelhttp.NewTransport(defaultTransport, otelhttp.WithSpanOptions(trace.WithAttributes(
			attribute.String("service", "minio"),
		))),
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}

	return raw, nil
}

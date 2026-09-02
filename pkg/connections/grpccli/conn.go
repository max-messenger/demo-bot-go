package grpccli

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/stats"
)

type Conn struct {
	config *Config

	dialOpts []grpc.DialOption

	conn *grpc.ClientConn

	healthCli grpc_health_v1.HealthClient
}

func NewConn(
	logger logging.Logger,
	config *Config,
) (*Conn, error) {
	return &Conn{
		config: config,
		dialOpts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithConnectParams(grpc.ConnectParams{
				Backoff:           config.Backoff.ToGRPC(),
				MinConnectTimeout: config.MinimumConnectTimeout,
			}),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler(
				otelgrpc.WithFilter(func(ri *stats.RPCTagInfo) bool {
					return ri.FullMethodName != grpc_health_v1.Health_Check_FullMethodName
				}),
			)),
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
			grpc.WithChainUnaryInterceptor(
				clMetrics.UnaryClientInterceptor(),
				logging.UnaryClientInterceptor(logger,
					logging.WithLogOnEvents(
						logging.FinishCall,
					),
					logging.WithLevels(defaultCodeToLevel),
				),
			),
			grpc.WithChainStreamInterceptor(
				clMetrics.StreamClientInterceptor(),
				logging.StreamClientInterceptor(logger,
					logging.WithLogOnEvents(
						logging.FinishCall,
					),
					logging.WithLevels(defaultCodeToLevel),
				),
			),
		},
	}, nil
}

func (c *Conn) RawConn() *grpc.ClientConn {
	return c.conn
}

func (c *Conn) Ping(ctx context.Context) error {
	if c.healthCli == nil {
		return nil
	}

	resp, err := c.healthCli.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "", // "" - overall status
	})

	if err != nil {
		return err
	}

	//nolint:exhaustive
	switch resp.Status {
	case grpc_health_v1.HealthCheckResponse_SERVING:
		return nil
	default:
		return fmt.Errorf("status %s", resp.Status.String())
	}
}

func (c *Conn) Start(_ context.Context) error {
	conn, err := grpc.NewClient(c.config.Addr, c.dialOpts...)
	if err != nil {
		return err
	}

	c.conn = conn
	c.healthCli = grpc_health_v1.NewHealthClient(conn)

	return nil
}

func (c *Conn) Stop(context.Context) error {
	if err := c.conn.Close(); err != nil {
		return err
	}

	c.conn = nil

	return nil
}

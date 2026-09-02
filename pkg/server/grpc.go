package server

import (
	"context"
	"fmt"
	"net"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"

	"demo_bot/pkg/telemetry"
)

var (
	grpcPanicsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "grpc_req_panics_recovered_total",
			Help: "Total number of gRPC requests recovered from internal panic.",
		},
	)

	gprcSrvMetrics = grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets(telemetry.DefaultHistogramBuckets),
		),
	)
)

func init() {
	prometheus.MustRegister(gprcSrvMetrics)
}

type GRPCParams struct {
	fx.In

	Logger *zap.Logger

	Config *GRPCConfig

	ServiceDefinitions []GRPCServiceDefinition `group:"impl_grpc"`
}
type GRPCServiceDefinition struct {
	Desc grpc.ServiceDesc
	Impl any
}

type GRPCServer struct {
	logger *zap.Logger
	config *GRPCConfig

	serviceDefinitions []GRPCServiceDefinition

	opts []grpc.ServerOption
	srv  *grpc.Server

	healthSrv *health.Server
}

func NewGRPCServer(
	params GRPCParams,
) (*GRPCServer, error) {
	if params.Config == nil {
		return nil, nil // nolint:nilnil  // it's ok
	}

	interceptorLogger := grpcInterceptorLogger(params.Logger)

	grpcPanicRecoveryOption := recovery.WithRecoveryHandler(func(p any) (err error) {
		grpcPanicsTotal.Inc()
		params.Logger.Error("recovered from panic", zap.Any("panic", p))

		return status.Errorf(codes.Internal, "%s", p)
	})

	return &GRPCServer{
		logger:             params.Logger,
		config:             params.Config,
		serviceDefinitions: params.ServiceDefinitions,
		opts: []grpc.ServerOption{
			grpc.StatsHandler(otelgrpc.NewServerHandler(
				otelgrpc.WithFilter(func(ri *stats.RPCTagInfo) bool {
					return ri.FullMethodName != grpc_health_v1.Health_Check_FullMethodName
				}),
			)),
			grpc.ChainUnaryInterceptor(
				gprcSrvMetrics.UnaryServerInterceptor(),
				logging.UnaryServerInterceptor(interceptorLogger,
					logging.WithLogOnEvents(
						logging.FinishCall,
					),
					logging.WithLevels(defaultCodeToLevel),
				),
				recovery.UnaryServerInterceptor(grpcPanicRecoveryOption),
			),
			grpc.ChainStreamInterceptor(
				gprcSrvMetrics.StreamServerInterceptor(),
				logging.StreamServerInterceptor(interceptorLogger,
					logging.WithLogOnEvents(
						logging.FinishCall,
					),
					logging.WithLevels(defaultCodeToLevel),
				),
				recovery.StreamServerInterceptor(grpcPanicRecoveryOption),
			),
		},
	}, nil
}

func (s *GRPCServer) Start(_ context.Context) error {
	ln, err := net.Listen("tcp", s.config.Addr) // nolint:noctx
	if err != nil {
		return err
	}

	s.logger.Info(
		"Starting GRPC server",
		zap.String("addr", s.config.Addr),
	)

	// create main grpc server with
	s.srv = grpc.NewServer(s.opts...)

	s.healthSrv = health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.srv, s.healthSrv)

	// register server definitions.
	for _, def := range s.serviceDefinitions {
		s.srv.RegisterService(&def.Desc, def.Impl)

	}

	reflection.Register(s.srv)

	go func() {
		if err := s.srv.Serve(ln); err != nil {
			s.logger.Error("Failed to start GRPC server", zap.Error(err))
		}
	}()

	s.healthSrv.SetServingStatus(
		"", grpc_health_v1.HealthCheckResponse_SERVING,
	)

	return nil
}

func (s *GRPCServer) Stop(_ context.Context) error {
	s.srv.Stop()

	return nil
}

func defaultCodeToLevel(code codes.Code) logging.Level {
	switch code {
	case codes.OK:
		return logging.LevelDebug

	case
		codes.NotFound,
		codes.Canceled,
		codes.AlreadyExists,
		codes.InvalidArgument,
		codes.Unauthenticated,
		codes.DeadlineExceeded,
		codes.PermissionDenied,
		codes.ResourceExhausted,
		codes.FailedPrecondition,
		codes.Aborted,
		codes.OutOfRange,
		codes.Unavailable:

		return logging.LevelWarn

	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return logging.LevelError

	default:
		return logging.LevelError
	}
}

func interceptorFields(fields []any) []zap.Field {
	f := make([]zap.Field, 0, len(fields)/2)

	for i := 0; i < len(fields); i += 2 {
		// nolint:errcheck
		key, _ := fields[i].(string)
		if key == "" {
			continue
		}

		value := fields[i+1]

		switch v := value.(type) {
		case string:
			f = append(f, zap.String(key, v))
		case int:
			f = append(f, zap.Int(key, v))
		case bool:
			f = append(f, zap.Bool(key, v))
		default:
			f = append(f, zap.Any(key, v))
		}
	}

	return f
}

func grpcInterceptorLogger(l *zap.Logger) logging.Logger { //nolint:ireturn
	return logging.LoggerFunc(func(
		_ context.Context, lvl logging.Level, msg string, fields ...any,
	) {
		logger := l.
			WithOptions(zap.AddCallerSkip(1)).
			With(interceptorFields(fields)...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			l.Error(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

package grpccli

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

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

func interceptorLogger(l *zap.Logger) logging.Logger { //nolint:ireturn
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

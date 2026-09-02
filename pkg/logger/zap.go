package logger

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logEnties = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_entries_total",
			Help: "Total number of log entries total",
		},
		[]string{"level"},
	)
)

func NewLogger(appName string) func(cfg Config) (*zap.Logger, error) {
	return func(cfg Config) (*zap.Logger, error) {
		lcfg := zap.NewProductionConfig()

		lcfg.Encoding = cfg.Format
		lcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		lcfg.DisableStacktrace = cfg.DisableStacktrace

		var lvl zapcore.Level
		if err := lvl.Set(cfg.Level); err != nil {
			return nil, fmt.Errorf("set zap level: %w", err)
		}

		lcfg.Level = zap.NewAtomicLevelAt(lvl)

		l, err := lcfg.Build()
		if err != nil {
			return nil, err
		}

		l = l.WithOptions(
			zap.Hooks(func(entry zapcore.Entry) error {
				logEnties.WithLabelValues(entry.Level.String()).Inc()

				return nil
			}),
		)

		return l.With(zap.String("app_name", appName)), nil
	}
}

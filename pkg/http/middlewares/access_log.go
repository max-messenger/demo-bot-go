package middlewares

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func AccessLog(logger *zap.Logger) func(next http.Handler) http.Handler {
	logger = logger.Named("access_log")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			chictx := chi.RouteContext(r.Context())
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			cost := time.Since(start)

			logFields := []zap.Field{
				zap.String("method", r.Method),
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.String("pattern", chictx.RoutePattern()),
				zap.Duration("cost", cost),
			}

			logger.Info(chictx.RoutePath, logFields...)
		})
	}
}

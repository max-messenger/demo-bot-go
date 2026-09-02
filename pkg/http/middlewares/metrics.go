package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "pattern", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds.",
			Buckets: prometheus.DefBuckets, // потом поменять
		},
		[]string{"method", "pattern"},
	)
)

func Metrics() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			chictx := chi.RouteContext(r.Context())
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			ts := time.Now()

			next.ServeHTTP(ww, r)

			cost := time.Since(ts)

			requestsTotal.WithLabelValues(
				r.Method,
				chictx.RoutePattern(),
				fmt.Sprint(ww.Status()),
			).Inc()

			requestDuration.WithLabelValues(
				r.Method,
				chictx.RoutePattern(),
			).Observe(cost.Seconds())
		})
	}

}

package middlewares

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.13.0/httpconv"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	headerRequestID = "X-Request-ID"
)

func Trace() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(headerRequestID)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			w.Header().Add(headerRequestID, requestID)

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			spanName := r.URL.Path
			chictx := chi.RouteContext(r.Context())
			if chictx.Routes.Match(chictx, r.Method, r.URL.Path) {
				spanName = chictx.RoutePattern()
			}
			chictx.Reset()

			attrs := httpconv.ServerRequest("", r)
			attrs = append(attrs, attribute.String(headerRequestID, requestID))

			ctx, span := otel.Tracer("http").Start(
				r.Context(),
				spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attrs...),
			)

			defer span.End()

			next.ServeHTTP(ww, r.WithContext(ctx))

			span.SetAttributes(
				semconv.HTTPResponseStatusCode(ww.Status()),
				semconv.HTTPResponseBodySize(ww.BytesWritten()),
			)
		})

	}
}

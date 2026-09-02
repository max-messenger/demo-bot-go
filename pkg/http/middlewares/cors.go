package middlewares

import (
	"net/http"
	"strings"
)

// CORSConfig конфигурация CORS.
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

// CORS возвращает middleware для обработки CORS-золовков.
// При пустом AllowedOrigins разрешаются все origins — origin запроса
// отражается в Access-Control-Allow-Origin с Allow-Credentials.
// При непустом — разрешается только указанный origin с Allow-Credentials.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	allowedSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowedSet[strings.ToLower(strings.TrimSpace(o))] = struct{}{}
	}

	allowAll := len(allowedSet) == 0

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)

				return
			}

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Vary", "Origin")

				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)

					return
				}

				next.ServeHTTP(w, r)

				return
			}

			if _, ok := allowedSet[strings.ToLower(origin)]; !ok {
				next.ServeHTTP(w, r)

				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Vary", "Origin")

			// Preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

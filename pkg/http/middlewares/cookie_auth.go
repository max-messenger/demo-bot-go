package middlewares

import (
	"context"
	"fmt"
	"net/http"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

// UserIDFromCtx извлекает user_id из контекста запроса.
func UserIDFromCtx(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(userIDKey).(int64)

	return uid, ok
}

// ParseTokenFunc — функция для валидации access-токена и извлечения user_id.
type ParseTokenFunc func(tokenString string) (userID int64, err error)

// CookieAuth возвращает middleware для проверки access_token куки.
func CookieAuth(parseToken ParseTokenFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "access token cookie not found")

				return
			}

			userID, err := parseToken(cookie.Value)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid access token")

				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := fmt.Fprintf(w, `{"error":%q}`, msg); err != nil {
		return
	}
}

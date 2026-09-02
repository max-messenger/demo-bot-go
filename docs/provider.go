package docs

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func Handler() http.HandlerFunc {
	router := chi.NewRouter()
	router.Get("/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return router.ServeHTTP
}

package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"demo_bot/pkg/http/middlewares"
)

// Controller api controller.
type Controller interface {
	Register(r chi.Router)
}

type Controllers []Controller

// 	@title          STG API
// 	@version        1.0
// 	@description    This is a demo_bot API.

//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func NewRouter(logger *zap.Logger, controllers Controllers) (http.Handler, error) {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Use(
			middlewares.Metrics(),
			middlewares.AccessLog(logger),
			middlewares.Trace(),
		)
		for _, ctrl := range controllers {
			ctrl.Register(r)
		}
	})

	return router, nil
}

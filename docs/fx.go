package docs

import (
	"net/http"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewOutAdapter,
	),
)

type OutAdapter struct {
	fx.Out

	SystemHTTPHandlers map[string]http.HandlerFunc `group:"system_handlers"`
}

func NewOutAdapter() OutAdapter {
	return OutAdapter{
		SystemHTTPHandlers: map[string]http.HandlerFunc{
			"/swagger": Handler(),
		},
	}
}

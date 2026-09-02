package services

import (
	"go.uber.org/fx"

	"demo_bot/internal/app/services/bot"
	"demo_bot/internal/app/services/state"
)

var Module = fx.Module(
	"services",
	fx.Options(
		state.Module,
		bot.Module,
	),
)

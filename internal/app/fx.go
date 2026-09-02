package app

import (
	"go.uber.org/fx"

	"demo_bot/docs"
	"demo_bot/internal/app/repository"
	"demo_bot/internal/app/router"
	"demo_bot/internal/app/services"
	"demo_bot/pkg/connections/maxbot"
	"demo_bot/pkg/connections/rediscli"
	"demo_bot/pkg/locker"
)

// Modules application modules.
var Modules = fx.Options(
	rediscli.Module,
	locker.Module,

	maxbot.Module,

	repository.Module,
	services.Module,

	router.Module,
	docs.Module,
)

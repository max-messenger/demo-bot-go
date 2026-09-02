package app

import (
	"go.uber.org/fx"

	"demo_bot"
)

func CreateApp(cfgName string) *fx.App {
	box := demo_bot.NewBox(
		Modules,
		demo_bot.WithAppName("demo_bot"),
		demo_bot.WithConfigFile(cfgName),
	)

	return box.CreateApp()
}

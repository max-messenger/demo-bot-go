package main

import (
	"flag"

	"demo_bot/internal/app"
)

func main() {
	var (
		cfgName string
	)

	flag.StringVar(&cfgName, "c", "config.yaml", "config")
	flag.Parse()

	app.CreateApp(cfgName).Run()
}

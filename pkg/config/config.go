package config

import (
	"os"

	"go.uber.org/config"
	"go.uber.org/fx"

	"demo_bot/pkg/bgtasker"
	"demo_bot/pkg/info"
	"demo_bot/pkg/logger"
	"demo_bot/pkg/server"
	"demo_bot/pkg/telemetry"
)

type Config struct {
	fx.Out

	App       AppConfig
	Logger    logger.Config
	Info      info.Config
	Telemetry telemetry.Config

	SystemServer server.SystemConfig
	HTTPServer   *server.HTTPConfig
	GRPCServer   *server.GRPCConfig

	Background *bgtasker.Config

	Provider config.Provider
}

type AppConfig struct {
	Hostname string `yaml:"hostname"`
}

type internalConfig struct {
	App       AppConfig        `yaml:"app"`
	Logger    logger.Config    `yaml:"logger"`
	Servers   server.Config    `yaml:"servers"`
	Telemetry telemetry.Config `yaml:"telemetry"`

	Background *bgtasker.Config `yaml:"background"`

	Provider config.Provider
}

func NewConfig(appName, cfgFile string) func() (Config, error) {
	return func() (Config, error) {
		inCfg, err := newInternalConfig(cfgFile)
		if err != nil {
			return Config{}, err
		}

		infoCfg := info.Config{
			AppName:  appName,
			Hostname: inCfg.App.Hostname,
		}
		inCfg.Telemetry.ServiceName = appName
		inCfg.Telemetry.Hostname = inCfg.App.Hostname
		inCfg.Telemetry.Version = infoCfg.BuildVersion()

		cfg := Config{
			App:    inCfg.App,
			Logger: inCfg.Logger,
			Info:   infoCfg,

			Telemetry: inCfg.Telemetry,

			Background: inCfg.Background,

			SystemServer: inCfg.Servers.System,
			HTTPServer:   inCfg.Servers.HTTP,
			GRPCServer:   inCfg.Servers.GRPC,

			// provider
			Provider: inCfg.Provider,
		}

		if inCfg.Background != nil {
			cfg.Background = inCfg.Background.Prepare()
		} else {
			cfg.Background = bgtasker.NewConfig()
		}

		return cfg, nil
	}
}

func newInternalConfig(fileName string) (internalConfig, error) {
	provider, err := config.NewYAML(
		config.Expand(os.LookupEnv),
		config.File(fileName),
		config.Permissive(),
	)

	if err != nil {
		return internalConfig{}, err
	}

	c := internalConfig{
		Logger:   logger.DefaultConfig(),
		Provider: provider,
	}

	err = provider.Get("").Populate(&c)
	if err != nil {
		return internalConfig{}, err
	}

	return c, nil
}

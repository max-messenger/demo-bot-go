package logger

type Config struct {
	Level             string `yaml:"level"`
	Format            string `yaml:"format"`
	DisableStacktrace bool   `yaml:"disable_stacktrace"`
}

func DefaultConfig() Config {
	return Config{
		Level:             "info",
		Format:            "json",
		DisableStacktrace: false,
	}
}

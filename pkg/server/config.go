package server

type Config struct {
	HTTP   *HTTPConfig  `yaml:"http"`
	GRPC   *GRPCConfig  `yaml:"grpc"`
	System SystemConfig `yaml:"system"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type GRPCConfig struct {
	Addr string `yaml:"addr"`
}

type SystemConfig struct {
	Addr string `yaml:"addr"`
}

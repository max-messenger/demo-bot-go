package ratelimiter

import "time"

type Config struct {
	Local  LocalConfig           `yaml:"local"`
	Shared SharedConfig          `yaml:"shared"`
	Rate   map[string]RateConfig `yaml:"rate"`
}

type RateConfig struct {
	Limit int                `yaml:"limit"`
	Keys  map[string]float64 `yaml:"keys"`
}

type LocalConfig struct {
	CacheSize int64                        `yaml:"cache_size"`
	TTL       time.Duration                `yaml:"ttl"`
	Custom    map[string]LocalCustomConfig `yaml:"custom"`
}

type LocalCustomConfig struct {
	CacheSize int64 `yaml:"cache_size"`
}

type SharedConfig struct {
	Prefix   string `yaml:"prefix"`
	PoolName string `yaml:"pool_name"`
}

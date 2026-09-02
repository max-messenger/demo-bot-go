package grpccli

import (
	"time"

	"google.golang.org/grpc/backoff"
)

const (
	defaultAddr                  = ":9000"
	defaultMinimumConnectTimeout = 3 * time.Second
)

type PoolConfig map[string]*Config

type BackoffConfig struct {
	BaseDelay  time.Duration `yaml:"base_delay"`
	Multiplier float64       `yaml:"multiplier"`
	Jitter     float64       `yaml:"jitter"`
	MaxDelay   time.Duration `yaml:"max_delay"`
}

func (c BackoffConfig) ToGRPC() backoff.Config {
	return backoff.Config{
		BaseDelay:  c.BaseDelay,
		Multiplier: c.Multiplier,
		Jitter:     c.Jitter,
		MaxDelay:   c.MaxDelay,
	}
}

type Config struct {
	Addr string `yaml:"addr"`

	MinimumConnectTimeout time.Duration `yaml:"minimum_connect_timeout"`

	Backoff BackoffConfig `yaml:"backoff"`
}

func (c *Config) Prepare() *Config {
	if c.Addr == "" {
		c.Addr = defaultAddr
	}

	if c.MinimumConnectTimeout == 0 {
		c.MinimumConnectTimeout = defaultMinimumConnectTimeout
	}

	if c.Backoff.BaseDelay == 0 {
		c.Backoff.BaseDelay = backoff.DefaultConfig.BaseDelay
	}

	if c.Backoff.Multiplier == 0 {
		c.Backoff.Multiplier = backoff.DefaultConfig.Multiplier
	}

	if c.Backoff.Jitter == 0 {
		c.Backoff.Jitter = backoff.DefaultConfig.Jitter
	}

	if c.Backoff.MaxDelay == 0 {
		c.Backoff.MaxDelay = backoff.DefaultConfig.MaxDelay
	}

	return c
}

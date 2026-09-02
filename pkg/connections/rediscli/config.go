package rediscli

import (
	"strings"
	"time"
)

type PoolConfig map[string]*Config

type Config struct {
	KeyPrefix string   `yaml:"key_prefix"`
	Addrs     []string `yaml:"addrs"`

	Username string `yaml:"username"`
	Password string `yaml:"password"` // nolint:gosec

	MasterName string `yaml:"master_name"`

	MaxRetries      int           `yaml:"max_retries"`
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`

	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`

	PoolSize        int           `yaml:"pool_size"`
	PoolTimeout     time.Duration `yaml:"pool_timeout"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxActiveConns  int           `yaml:"max_active_conns"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`

	MaxRedirects   int  `yaml:"max_redirects"`
	RouteByLatency bool `yaml:"route_by_latency"`
	RouteRandomly  bool `yaml:"route_randomly"`
}

func (c *Config) Prepare() *Config {
	if len(c.Addrs) == 1 {
		c.Addrs = strings.Split(c.Addrs[0], ",")
	}

	return c
}

package locker

import "time"

const (
	defaultTTL     = 10 * time.Second
	defaultRetries = 10
	defaultDelay   = 200 * time.Millisecond
)

type Config struct {
	PoolName string        `yaml:"pool_name"`
	TTL      time.Duration `yaml:"ttl"`
	Retries  int           `yaml:"retries"`
	Delay    time.Duration `yaml:"delay"`
}

func NewConfig() Config {
	return Config{
		TTL:     defaultTTL,
		Retries: defaultRetries,
		Delay:   defaultDelay,
	}
}

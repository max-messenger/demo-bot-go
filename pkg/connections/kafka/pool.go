package kafka

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

var (
	// ErrPoolNotFound happens when try to get non existing pool.
	ErrPoolNotFound = errors.New("pool not found")
)

type Pool struct {
	cfgs   map[string]*Config
	pools  map[string]*Kafka
	logger *zap.Logger
}

func NewPool(
	logger *zap.Logger,
	cfgs PoolConfig,
) (*Pool, error) {
	p := &Pool{
		cfgs:  cfgs,
		pools: make(map[string]*Kafka, len(cfgs)),

		logger: logger,
	}

	for name, cfg := range p.cfgs {
		kf, err := NewKafka(
			logger.Named(name),
			cfg,
			name,
		)
		if err != nil {
			return nil, fmt.Errorf("create kafka cli %s: %w", name, err)
		}
		p.pools[name] = kf
	}

	return p, nil
}

func (p *Pool) GetPool(name string) (*Kafka, error) {
	if kf, ok := p.pools[name]; ok {
		return kf, nil
	}

	return nil, ErrPoolNotFound
}

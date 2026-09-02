package rediscli

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

var (
	// ErrPoolNotFound happens when try to get non existing pool.
	ErrPoolNotFound = errors.New("pool not found")
)

type Pool struct {
	pools map[string]*Redis
	mu    sync.RWMutex

	logger *zap.Logger
}

func NewPool(
	logger *zap.Logger,
	cfgs PoolConfig,
) (*Pool, error) {
	p := &Pool{

		pools: make(map[string]*Redis, len(cfgs)),

		logger: logger,
	}

	for name, cfg := range cfgs {
		redis, err := NewRedis(
			logger,
			cfg,
			name,
		)

		if err != nil {
			return nil, err
		}

		p.pools[name] = redis
	}

	return p, nil
}

func (p *Pool) GetPool(name string) (*Redis, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if pool, ok := p.pools[name]; ok {
		return pool, nil
	}

	return nil, ErrPoolNotFound
}

func (p *Pool) Stop(_ context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, pool := range p.pools {
		err := pool.Close()
		if err != nil {
			p.logger.Error("stop redis", zap.Error(err))

			return err
		}
	}

	return nil
}

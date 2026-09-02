package grpccli

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var (
	// ErrConnNotFound happens when try to get non existing connection.
	ErrConnNotFound = errors.New("conn not found")
)

type Pool struct {
	logger      *zap.Logger
	configMap   map[string]*Config
	connections map[string]*Conn

	mu sync.RWMutex
}

func NewPool(
	logger *zap.Logger,
	configMap map[string]*Config,
) (*Pool, error) {
	p := &Pool{
		logger:      logger,
		configMap:   configMap,
		connections: make(map[string]*Conn, len(configMap)),
	}

	interceptorLogger := interceptorLogger(p.logger)

	for name, cfg := range p.configMap {
		conn, err := NewConn(
			interceptorLogger,
			cfg,
		)

		if err != nil {
			return nil, err
		}

		err = conn.Start(context.Background())
		if err != nil {
			return nil, err
		}

		p.connections[name] = conn
	}

	return p, nil
}

func (p *Pool) GetConn(name string) (*Conn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if pool, ok := p.connections[name]; ok {
		return pool, nil
	}

	return nil, fmt.Errorf("%w: conn. name: %s", ErrConnNotFound, name)
}

func (p *Pool) GetOrCreateConn(
	ctx context.Context,
	name string,
	cfg *Config,
) (*Conn, error) {

	p.mu.RLock()
	pool, ok := p.connections[name]
	p.mu.RUnlock()

	if ok {
		return pool, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	pool, ok = p.connections[name]
	if ok {
		return pool, nil
	}

	conn, err := NewConn(
		interceptorLogger(p.logger),
		cfg,
	)

	if err != nil {
		return nil, fmt.Errorf("new connect: %w", err)
	}

	err = conn.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("start connect: %w", err)
	}

	p.configMap[name] = cfg
	p.connections[name] = conn

	return conn, nil
}

func (p *Pool) Stop(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, pool := range p.connections {
		if err := pool.Stop(ctx); err != nil {
			return fmt.Errorf("stop connection: %w", err)
		}
	}

	return nil
}

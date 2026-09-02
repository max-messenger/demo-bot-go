package bgtasker

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

type Pool struct {
	logger *zap.Logger

	taskers map[string]*Runner

	m       sync.RWMutex
	started bool
}

func NewPool(logger *zap.Logger, config *Config) *Pool {
	p := &Pool{
		logger: logger,

		taskers: make(map[string]*Runner),
	}

	for name, taskerConfig := range config.Tasks {
		_, err := p.newTasker(name, taskerConfig)
		if err != nil {
			return nil
		}
	}

	return p
}

func (p *Pool) Start(ctx context.Context) error {
	p.m.Lock()
	defer p.m.Unlock()

	for _, tasker := range p.taskers {
		if err := tasker.Start(ctx); err != nil {
			return err
		}
	}

	p.started = true

	return nil
}

func (p *Pool) Stop(ctx context.Context) error {
	p.m.Lock()
	defer p.m.Unlock()

	for _, tasker := range p.taskers {
		if err := tasker.Stop(ctx); err != nil {
			return err
		}
	}

	p.started = false

	return nil
}

func (p *Pool) Get(name string) (*Runner, error) {
	p.m.RLock()
	defer p.m.RUnlock()

	tasker, ok := p.taskers[name]
	if !ok {
		t, err := p.newTasker(name, DefaultTaskConfig())
		if err != nil {
			return nil, err
		}

		tasker = t

		p.logger.Warn("using default task config", zap.String("name", name))
	}

	return tasker, nil
}

func (p *Pool) newTasker(name string, config *TaskConfig) (*Runner, error) {
	if p.started {
		return nil, errors.New("pool already started")
	}

	tasker := NewTasker(
		p.logger.Named(name),
		config,
		name,
	)

	p.taskers[name] = tasker

	return tasker, nil
}

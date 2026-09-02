package grace

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Service graceful service interface.
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ServicePool service pool for graceful shutdown services.
type ServicePool struct {
	services []Service
	logger   *zap.Logger
}

func NewServicePool(pi Params) *ServicePool {
	return &ServicePool{
		services: pi.Services,
		logger:   pi.Logger,
	}
}

// Start starting services.
func (g *ServicePool) Start(ctx context.Context) error {
	g.logger.Info("Service Pool: run services")

	for _, s := range g.services {
		if err := s.Start(ctx); err != nil {
			return fmt.Errorf("start service: %w", err)
		}

		g.logger.Info("Service Pool: service started",
			zap.String("service", fmt.Sprintf("%T", s)),
		)
	}

	g.logger.Info("Service Pool: All Service started 🚀")

	return nil
}

// Stop stoping services.
func (g *ServicePool) Stop(ctx context.Context) error {
	g.logger.Info("Service Pool: Stopping services")
	var wg sync.WaitGroup

	for _, s := range g.services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()

			err := s.Stop(ctx)
			if err != nil {
				g.logger.Error("Service pool: stop service error",
					zap.String("service", fmt.Sprintf("%T", s)),
					zap.Error(err),
				)

				return
			}

			g.logger.Info("Service pool: service stopped",
				zap.String("service", fmt.Sprintf("%T", s)),
			)
		}(s)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		g.logger.Info("Service pool: All services stoped 🪦")
	case <-ctx.Done():
		g.logger.Error("Service pool: Stop aborted by context 💀")
	}

	return nil
}

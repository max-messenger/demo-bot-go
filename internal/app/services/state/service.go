package state

import (
	"context"

	"demo_bot/internal/app/domain"
)

// Service manages scenario state lifecycle.
type Service interface {
	GetState(ctx context.Context, userID int64) (*domain.ScenarioState, error)
	SetState(ctx context.Context, userID int64, state *domain.ScenarioState) error
	ClearState(ctx context.Context, userID int64) error
}

//go:generate go tool mockgen -source=service.go -destination=mocks/service_mocks.go -package=mocks

type service struct {
	repo Repository
}

// Repository defines the data access interface for state storage.
type Repository interface {
	GetState(ctx context.Context, userID int64) (*domain.ScenarioState, error)
	SetState(ctx context.Context, userID int64, state *domain.ScenarioState) error
	ClearState(ctx context.Context, userID int64) error
}

// NewService creates a new state service.
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetState(ctx context.Context, userID int64) (*domain.ScenarioState, error) {
	return s.repo.GetState(ctx, userID)
}

func (s *service) SetState(ctx context.Context, userID int64, state *domain.ScenarioState) error {
	return s.repo.SetState(ctx, userID, state)
}

func (s *service) ClearState(ctx context.Context, userID int64) error {
	return s.repo.ClearState(ctx, userID)
}

package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"demo_bot/internal/app/domain"
	"demo_bot/pkg/connections/rediscli"
)

const (
	domainKey  = "state"
	traceGet   = "state-get"
	traceSet   = "state-set"
	traceClear = "state-clear"
	defaultTTL = 30 * time.Minute
)

// Store abstracts the Redis call mechanism for tracing and key formatting.
type Store interface {
	Call(ctx context.Context, target string, f rediscli.CallCallback) error
}

// Repository implements Redis-backed scenario state storage.
type Repository struct {
	store Store
	ttl   time.Duration
}

// NewRepository creates a new state repository.
func NewRepository(store Store, cfg Config) *Repository {
	ttl := defaultTTL
	if cfg.TTL > 0 {
		ttl = cfg.TTL
	}

	return &Repository{
		store: store,
		ttl:   ttl,
	}
}

// GetState retrieves the current scenario state for a user.
// Returns nil if no state exists.
func (r *Repository) GetState(ctx context.Context, userID int64) (*domain.ScenarioState, error) {
	var state *domain.ScenarioState

	err := r.store.Call(ctx, traceGet, func(
		ctx context.Context,
		ucl redis.UniversalClient,
		kf rediscli.KeyFormatter,
	) error {
		val, err := ucl.Get(ctx, kf.FormatKey(r.serializeKey(userID))).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil
			}

			return fmt.Errorf("get state: %w", err)
		}

		state = &domain.ScenarioState{}

		return json.Unmarshal([]byte(val), state)
	})
	if err != nil {
		return nil, fmt.Errorf("state repository get: %w", err)
	}

	return state, nil
}

// SetState saves the scenario state for a user with TTL.
func (r *Repository) SetState(ctx context.Context, userID int64, state *domain.ScenarioState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	err = r.store.Call(ctx, traceSet, func(
		ctx context.Context,
		ucl redis.UniversalClient,
		kf rediscli.KeyFormatter,
	) error {
		return ucl.Set(ctx, kf.FormatKey(r.serializeKey(userID)), data, r.ttl).Err()
	})
	if err != nil {
		return fmt.Errorf("state repository set: %w", err)
	}

	return nil
}

// ClearState removes the scenario state for a user.
func (r *Repository) ClearState(ctx context.Context, userID int64) error {
	err := r.store.Call(ctx, traceClear, func(
		ctx context.Context,
		ucl redis.UniversalClient,
		kf rediscli.KeyFormatter,
	) error {
		return ucl.Del(ctx, kf.FormatKey(r.serializeKey(userID))).Err()
	})
	if err != nil {
		return fmt.Errorf("state repository clear: %w", err)
	}

	return nil
}

func (r *Repository) serializeKey(userID int64) string {
	return fmt.Sprintf("%s:%d", domainKey, userID)
}

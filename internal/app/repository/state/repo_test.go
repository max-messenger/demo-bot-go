package state_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"demo_bot/internal/app/domain"
	repostate "demo_bot/internal/app/repository/state"
	"demo_bot/pkg/connections/rediscli"
)

const testUserID int64 = 42

type testStore struct {
	client redis.UniversalClient
}

func (s *testStore) Call(
	ctx context.Context,
	_ string,
	f rediscli.CallCallback,
) error {
	return f(ctx, s.client, testKeyFormatter{})
}

type testKeyFormatter struct{}

func (testKeyFormatter) FormatKey(key string) string {
	return "test:" + key
}

func newTestRepo(t *testing.T) (*repostate.Repository, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{mr.Addr()},
	})
	t.Cleanup(func() { _ = client.Close() })

	store := &testStore{client: client}
	repo := repostate.NewRepository(store, repostate.Config{})

	return repo, mr
}

func newTestRepoWithTTL(
	t *testing.T,
	ttl time.Duration,
) (*repostate.Repository, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{mr.Addr()},
	})
	t.Cleanup(func() { _ = client.Close() })

	store := &testStore{client: client}
	repo := repostate.NewRepository(store, repostate.Config{TTL: ttl})

	return repo, mr
}

func TestUnitRepository_NewRepository_DefaultTTL(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)

	ctx := context.Background()
	state := &domain.ScenarioState{
		Scenario: "greeting",
		Step:     1,
		Data:     map[string]any{"key": "value"},
	}

	err := repo.SetState(ctx, testUserID, state)
	require.NoError(t, err)

	// Immediately after set, state should be available.
	got, err := repo.GetState(ctx, testUserID)
	require.NoError(t, err)
	assert.Equal(t, state, got)
}

func TestUnitRepository_NewRepository_CustomTTL(t *testing.T) {
	t.Parallel()

	const customTTL = 100 * time.Millisecond

	repo, mr := newTestRepoWithTTL(t, customTTL)

	ctx := context.Background()
	state := &domain.ScenarioState{
		Scenario: "onboarding",
		Step:     2,
	}

	err := repo.SetState(ctx, testUserID, state)
	require.NoError(t, err)

	// Before TTL expires, state should exist.
	got, err := repo.GetState(ctx, testUserID)
	require.NoError(t, err)
	assert.Equal(t, state, got)

	// Fast-forward past TTL.
	mr.FastForward(customTTL + time.Millisecond)

	// After TTL, state should be gone (nil, no error).
	got, err = repo.GetState(ctx, testUserID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestUnitRepository_GetState_Success(t *testing.T) {
	t.Parallel()

	repo, mr := newTestRepo(t)

	expected := &domain.ScenarioState{
		Scenario: "survey",
		Step:     3,
		Data:     map[string]any{"answer": "yes"},
	}

	ctx := context.Background()

	err := repo.SetState(ctx, testUserID, expected)
	require.NoError(t, err)

	got, err := repo.GetState(ctx, testUserID)
	require.NoError(t, err)
	assert.Equal(t, expected, got)

	// Verify the key is formatted correctly in Redis.
	key := "test:state:" + formatUserID(testUserID)
	assert.True(t, mr.Exists(key))
}

func TestUnitRepository_GetState_NotFound(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)

	ctx := context.Background()

	got, err := repo.GetState(ctx, testUserID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestUnitRepository_GetState_InvalidJSON(t *testing.T) {
	t.Parallel()

	repo, mr := newTestRepo(t)

	ctx := context.Background()

	// Write invalid JSON directly into miniredis.
	key := "test:state:" + formatUserID(testUserID)
	mr.Set(key, "{invalid-json")

	got, err := repo.GetState(ctx, testUserID)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "state repository get")
}

func TestUnitRepository_SetState_Success(t *testing.T) {
	t.Parallel()

	repo, mr := newTestRepo(t)

	ctx := context.Background()

	state := &domain.ScenarioState{
		Scenario: "support",
		Step:     1,
		Data:     map[string]any{"ticket": float64(99)},
	}

	err := repo.SetState(ctx, testUserID, state)
	require.NoError(t, err)

	// Verify data stored in Redis.
	key := "test:state:" + formatUserID(testUserID)
	val, err := mr.Get(key)
	require.NoError(t, err)

	var stored domain.ScenarioState
	require.NoError(t, json.Unmarshal([]byte(val), &stored))
	assert.Equal(t, *state, stored)
}

func TestUnitRepository_ClearState_Success(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)

	ctx := context.Background()

	state := &domain.ScenarioState{
		Scenario: "farewell",
		Step:     1,
	}

	err := repo.SetState(ctx, testUserID, state)
	require.NoError(t, err)

	err = repo.ClearState(ctx, testUserID)
	require.NoError(t, err)

	// After clearing, GetState should return nil.
	got, err := repo.GetState(ctx, testUserID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestUnitRepository_ClearState_NotExisting(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)

	ctx := context.Background()

	// Clearing a non-existing key should not return an error.
	err := repo.ClearState(ctx, testUserID)
	require.NoError(t, err)
}

func formatUserID(id int64) string {
	return fmt.Sprintf("%d", id)
}

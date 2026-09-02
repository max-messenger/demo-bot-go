package scenario_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"demo_bot/internal/app/services/bot/scenario"
)

// testScenario is a minimal Scenario implementation for registry tests.
type testScenario struct {
	name        string
	description string
	chatType    scenario.ChatType
	permissions []string
}

func (s testScenario) Name() string                { return s.name }
func (s testScenario) Description() string         { return s.description }
func (s testScenario) ChatType() scenario.ChatType { return s.chatType }

func (s testScenario) RequiredPermissions() []string { return s.permissions }

func (s testScenario) Handle(
	_ context.Context,
	_ scenario.Params,
	st *scenario.State,
) (*scenario.State, error) {
	return st, nil
}

func newTestScenario(name string, chatType scenario.ChatType) testScenario {
	return testScenario{
		name:        name,
		description: "test scenario " + name,
		chatType:    chatType,
		permissions: nil,
	}
}

func TestUnitRegistry_AddAndGet(t *testing.T) {
	t.Parallel()

	r := scenario.NewRegistry()
	s := newTestScenario("message", scenario.ChatTypePersonal)

	r.Add(s)

	got := r.Get("message")
	require.NotNil(t, got)
	assert.Equal(t, "message", got.Name())
	assert.Equal(t, scenario.ChatTypePersonal, got.ChatType())
}

func TestUnitRegistry_GetNotFound(t *testing.T) {
	t.Parallel()

	r := scenario.NewRegistry()

	got := r.Get("nonexistent")
	assert.Nil(t, got)
}

func TestUnitRegistry_AddDuplicatePanics(t *testing.T) {
	t.Parallel()

	r := scenario.NewRegistry()
	s := newTestScenario("duplicate", scenario.ChatTypePersonal)

	r.Add(s)

	assert.Panics(t, func() {
		r.Add(newTestScenario("duplicate", scenario.ChatTypeGroup))
	})
}

func TestUnitRegistry_ListByChatType(t *testing.T) {
	t.Parallel()

	r := scenario.NewRegistry()
	r.Add(newTestScenario("attachments", scenario.ChatTypeAny))
	r.Add(newTestScenario("actions", scenario.ChatTypeAny))
	r.Add(newTestScenario("mention", scenario.ChatTypeGroup))
	r.Add(newTestScenario("message", scenario.ChatTypePersonal))
	r.Add(newTestScenario("links", scenario.ChatTypePersonal))

	t.Run("personal includes any and personal", func(t *testing.T) {
		t.Parallel()

		result := r.ListByChatType(scenario.ChatTypePersonal)
		names := scenarioNames(t, result)

		assert.Equal(t, []string{"attachments", "actions", "links", "message"}, names)
	})

	t.Run("group includes any and group", func(t *testing.T) {
		t.Parallel()

		result := r.ListByChatType(scenario.ChatTypeGroup)
		names := scenarioNames(t, result)

		assert.Equal(t, []string{"attachments", "actions", "mention"}, names)
	})

	t.Run("empty registry returns nil", func(t *testing.T) {
		t.Parallel()

		empty := scenario.NewRegistry()
		result := empty.ListByChatType(scenario.ChatTypePersonal)

		assert.Nil(t, result)
	})
}

func TestUnitRegistry_SortingOrder(t *testing.T) {
	t.Parallel()

	r := scenario.NewRegistry()

	// Add scenarios in non-sorted order; scenarioOrder defines:
	// attachments=1, actions=2, mention=3, rest alphabetically.
	r.Add(newTestScenario("zebra", scenario.ChatTypeAny))
	r.Add(newTestScenario("mention", scenario.ChatTypeAny))
	r.Add(newTestScenario("actions", scenario.ChatTypeAny))
	r.Add(newTestScenario("attachments", scenario.ChatTypeAny))
	r.Add(newTestScenario("alpha", scenario.ChatTypeAny))

	result := r.ListByChatType(scenario.ChatTypePersonal)
	names := scenarioNames(t, result)

	// Ordered scenarios first (attachments, actions, mention), then alphabetical.
	assert.Equal(t, []string{"attachments", "actions", "mention", "alpha", "zebra"}, names)
}

// scenarioNames extracts scenario names for easier assertion.
func scenarioNames(t *testing.T, scenarios []scenario.Scenario) []string {
	t.Helper()

	names := make([]string, len(scenarios))
	for i, s := range scenarios {
		names[i] = s.Name()
	}

	return names
}

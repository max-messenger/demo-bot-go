package scenario

import "sort"

// scenarioOrder defines explicit display order for scenarios in the start menu.
// Scenarios not in this map sort alphabetically after all ordered ones.
var scenarioOrder = map[string]int{
	"attachments": 1,
	"actions":     2,
	"mention":     3,
}

// Registry holds all registered scenarios, indexed by name.
type Registry struct {
	scenarios map[string]Scenario
}

// NewRegistry creates a new empty scenario registry.
func NewRegistry() *Registry {
	return &Registry{
		scenarios: make(map[string]Scenario),
	}
}

// Add registers a scenario. Panics if a scenario with the same name already exists.
func (r *Registry) Add(s Scenario) {
	if _, exists := r.scenarios[s.Name()]; exists {
		panic("scenario already registered: " + s.Name())
	}

	r.scenarios[s.Name()] = s
}

// Get returns a scenario by name, or nil if not found.
func (r *Registry) Get(name string) Scenario {
	return r.scenarios[name]
}

// ListByChatType returns all scenarios matching the given chat type.
// Scenarios with ChatTypeAny are included in both personal and group results.
func (r *Registry) ListByChatType(chatType ChatType) []Scenario {
	var result []Scenario

	for _, s := range r.scenarios {
		if s.ChatType() == chatType || s.ChatType() == ChatTypeAny {
			result = append(result, s)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		ni, nj := result[i].Name(), result[j].Name()
		oi, okI := scenarioOrder[ni]
		oj, okJ := scenarioOrder[nj]

		switch {
		case okI && okJ:
			return oi < oj
		case okI:
			return true
		case okJ:
			return false
		default:
			return ni < nj
		}
	})

	return result
}

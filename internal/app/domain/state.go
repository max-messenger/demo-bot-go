package domain

// ChatType defines the type of chat where a scenario can run.
type ChatType string

const (
	// ChatTypePersonal means the scenario runs in direct messages with the bot.
	ChatTypePersonal ChatType = "personal"
	// ChatTypeGroup means the scenario runs in group chats.
	ChatTypeGroup ChatType = "group"
	// ChatTypeAny means the scenario can run in any chat type.
	ChatTypeAny ChatType = "any"
)

// ScenarioState holds the current state of an active scenario for a user.
type ScenarioState struct {
	Scenario string         `json:"scenario"`
	Step     int            `json:"step"`
	Data     map[string]any `json:"data,omitempty"`
}

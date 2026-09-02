package scenario

import (
	"context"
	"errors"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
	"go.uber.org/zap"

	"demo_bot/internal/app/domain"
)

// API defines the bot API client interface.
type API interface {
	Client() *maxbotcli.Api
}

//go:generate go tool mockgen -source=scenario.go -destination=mocks/scenario_mocks.go -package=mocks

// ErrDone is returned by Scenario.Handle when the scenario is complete.
// The caller should clear the user's state when this error is received.
var ErrDone = errors.New("scenario completed")

// ErrDoneNoMenu is returned when the scenario is complete but the start menu
// should NOT be sent (e.g., the bot left the chat and cannot send messages).
var ErrDoneNoMenu = errors.New("scenario completed, skip menu")

// ChatType defines where a scenario can run.
type ChatType = domain.ChatType

const (
	ChatTypePersonal = domain.ChatTypePersonal
	ChatTypeGroup    = domain.ChatTypeGroup
	ChatTypeAny      = domain.ChatTypeAny
)

// State holds the current step and data for an active scenario.
type State = domain.ScenarioState

// Params carries dependencies available to scenario handlers.
type Params struct {
	API API
	Ctx maxbot.Context
	Log *zap.Logger
}

// Scenario is the interface every demo scenario must implement.
type Scenario interface {
	// Name returns the scenario command name (e.g. "message" for /message).
	Name() string
	// Description returns a short Russian description for /help.
	Description() string
	// ChatType returns where this scenario can run.
	ChatType() ChatType
	// RequiredPermissions returns admin permissions needed (group only).
	RequiredPermissions() []string
	// Handle processes one step of the scenario.
	// Step 0 means the scenario is starting.
	// Return non-nil State to continue, nil to complete.
	Handle(ctx context.Context, p Params, s *State) (*State, error)
}

// Store defines the state persistence interface.
type Store interface {
	GetState(ctx context.Context, userID int64) (*State, error)
	SetState(ctx context.Context, userID int64, state *State) error
	ClearState(ctx context.Context, userID int64) error
}

// IsPersonalChat checks if the update comes from a personal (dialog) chat.
// For updates without Message (e.g., bot_started), defaults to personal.
func IsPersonalChat(upd model.Update) bool {
	if upd.Message != nil {
		return upd.Message.Recipient.ChatType == model.ChatTypeDialog
	}

	return true
}

// IsGroupChat checks if the update comes from a group chat.
func IsGroupChat(upd model.Update) bool {
	if upd.Message != nil {
		return upd.Message.Recipient.ChatType == model.ChatTypeChat
	}

	return false
}

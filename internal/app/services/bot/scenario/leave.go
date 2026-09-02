package scenario

import (
	"context"
	"fmt"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
)

const (
	leaveYes = "leave:yes"
	leaveNo  = "leave:no"
)

// LeaveScenario demonstrates leaving a group chat.
type LeaveScenario struct{}

// NewLeaveScenario creates a new leave demo scenario.
func NewLeaveScenario() *LeaveScenario {
	return &LeaveScenario{}
}

// Name returns the scenario command name.
func (LeaveScenario) Name() string { return "leave" }

// Description returns a short Russian description for /help.
func (LeaveScenario) Description() string {
	return "демо выхода из чата"
}

// ChatType returns Group — this scenario only works in group chats.
func (LeaveScenario) ChatType() ChatType { return ChatTypeGroup }

// RequiredPermissions returns admin permissions needed for the leave scenario.
func (LeaveScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the leave demo scenario.
func (l LeaveScenario) Handle(ctx context.Context, p Params, _ *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	// Check for callback with leave decision.
	if upd.Callback != nil {
		switch upd.Callback.Payload {
		case leaveYes:
			if _, err := p.API.Client().Chats.LeaveChat(ctx, chatID); err != nil {
				return nil, fmt.Errorf("leave chat: %w", err)
			}

			return nil, ErrDoneNoMenu
		case leaveNo:
			if err := p.Ctx.Send("Отменено"); err != nil {
				return nil, fmt.Errorf("send cancel message: %w", err)
			}

			return nil, ErrDone
		default:
			// Unknown callback, ignore and re-show confirmation.
		}
	}

	// Step 0: show confirmation with "Да"/"Нет" buttons.
	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("Да", leaveYes)
	kb.AddRow().AddCallBack("Нет", leaveNo)

	if err := p.Ctx.Send("Покинуть чат?", maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send leave confirmation: %w", err)
	}

	return &State{Scenario: l.Name(), Step: 0}, nil
}

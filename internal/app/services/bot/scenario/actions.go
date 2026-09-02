package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
)

// sleepWithContext pauses for the given duration or until the context is cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

const (
	payloadActRepeat = "act:repeat"
	payloadActNext   = "act:next"
)

// actionStep defines a single step in the actions demo: an action to send and text to explain it.
type actionStep struct {
	action model.SenderAction
	text   string
}

var actionSteps = []actionStep{
	{action: model.ActionTypingOn, text: "Индикатор ввода отправлен"},
	{action: model.ActionSendingPhoto, text: "Индикатор отправки фото"},
	{action: model.ActionSendingVideo, text: "Индикатор отправки видео"},
	{action: model.ActionSendingAudio, text: "Индикатор отправки аудио"},
	{action: model.ActionSendingFile, text: "Индикатор отправки файла"},
	{action: model.ActionMarkSeen, text: "Демо действий завершено!"},
}

// ActionsScenario demonstrates all chat action indicators via Chats.SendAction.
type ActionsScenario struct{}

// NewActionsScenario creates a new actions scenario.
func NewActionsScenario() *ActionsScenario {
	return &ActionsScenario{}
}

// Name returns the scenario command name.
func (ActionsScenario) Name() string { return "actions" }

// Description returns a short Russian description for /help.
func (ActionsScenario) Description() string {
	return "демо индикаторов действий в чате"
}

// ChatType returns where this scenario can run.
func (ActionsScenario) ChatType() ChatType { return ChatTypePersonal }

// RequiredPermissions returns admin permissions needed (group only).
func (ActionsScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the actions scenario.
// Step 0 is the intro message. Steps 1-6 are action steps (actionIdx = step - 1).
func (a ActionsScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()

	// Step 0: show intro message, wait for "Начать" click.
	if s.Step == 0 {
		if upd.Callback != nil && upd.Callback.Payload == payloadActNext {
			// User clicked "Начать" — execute first action directly.
			cur := actionSteps[0]

			if _, err := p.API.Client().Chats.SendAction(ctx, upd.ChatID, cur.action); err != nil {
				return nil, fmt.Errorf("send action %s: %w", cur.action, err)
			}

			if err := sleepWithContext(ctx, 1500*time.Millisecond); err != nil {
				return nil, fmt.Errorf("action delay: %w", err)
			}

			return a.sendStepResult(p, cur, 0)
		}

		if upd.Callback != nil && upd.Callback.Payload == PayloadCancel {
			return nil, ErrDone
		}

		// Show intro (first visit, unknown callback, or text message).
		kb := model.NewKeyboard()
		kb.AddRow().AddCallBack("Начать", payloadActNext)
		kb.AddRow().AddCallBack("Завершить", PayloadCancel)

		if err := p.Ctx.Send(
			"Сейчас я буду отправлять индикаторы действий в чате (Chats.SendAction). "+
				"Обратите внимание на строку состояния чата — там появятся анимации: «печатает», «отправляет фото» и т.д.",
			maxbot.WithKeyboard(kb),
		); err != nil {
			return nil, fmt.Errorf("send intro: %w", err)
		}

		return &State{Scenario: a.Name(), Step: 0}, nil
	}

	// Steps 1+: resolve callback/text and execute action.
	return a.handleAction(ctx, p, upd, s.Step)
}

// handleAction resolves the next action index and sends the action.
func (a ActionsScenario) handleAction(ctx context.Context, p Params, upd model.Update, currentStep int) (*State, error) {
	step, err := a.resolveActionStep(upd, p, currentStep)
	if err != nil {
		return nil, err
	}

	if step < 0 {
		return nil, ErrDone
	}

	if a.isTextHint(step, currentStep, upd) {
		return &State{Scenario: a.Name(), Step: currentStep}, nil
	}

	return a.executeAction(ctx, p, upd, step)
}

func (a ActionsScenario) isTextHint(step, currentStep int, upd model.Update) bool {
	return step == currentStep && upd.Callback == nil && upd.Message != nil && upd.Message.Body.Text != ""
}

func (a ActionsScenario) executeAction(ctx context.Context, p Params, upd model.Update, step int) (*State, error) {
	actionIdx := step - 1
	if actionIdx < 0 || actionIdx >= len(actionSteps) {
		return nil, fmt.Errorf("actions: unknown step %d", step)
	}

	cur := actionSteps[actionIdx]

	if _, err := p.API.Client().Chats.SendAction(ctx, upd.ChatID, cur.action); err != nil {
		return nil, fmt.Errorf("send action %s: %w", cur.action, err)
	}

	if err := sleepWithContext(ctx, 1500*time.Millisecond); err != nil {
		return nil, fmt.Errorf("action delay: %w", err)
	}

	return a.sendStepResult(p, cur, actionIdx)
}

func (a ActionsScenario) resolveActionStep(upd model.Update, p Params, currentStep int) (int, error) {
	if upd.Callback != nil {
		switch upd.Callback.Payload {
		case payloadActNext:
			return currentStep + 1, nil
		case payloadActRepeat:
			return currentStep, nil
		case PayloadCancel:
			return -1, nil
		default:
			return currentStep, nil
		}
	}

	if upd.Message != nil && upd.Message.Body.Text != "" {
		if err := p.Ctx.Send("Пожалуйста, используйте кнопки «Повторить» или «Далее» ниже."); err != nil {
			return 0, fmt.Errorf("send hint: %w", err)
		}

		return currentStep, nil
	}

	return currentStep, nil
}

func (a ActionsScenario) sendStepResult(p Params, cur actionStep, actionIdx int) (*State, error) {
	if actionIdx == len(actionSteps)-1 {
		if err := p.Ctx.Send(cur.text); err != nil {
			return nil, fmt.Errorf("send final message: %w", err)
		}

		return nil, ErrDone
	}

	keyboard := model.NewKeyboard()
	keyboard.AddRow().AddCallBack("Повторить", payloadActRepeat)
	keyboard.AddRow().AddCallBack("Далее", payloadActNext)
	keyboard.AddRow().AddCallBack("Завершить", PayloadCancel)

	if err := p.Ctx.Send(cur.text, maxbot.WithKeyboard(keyboard)); err != nil {
		return nil, fmt.Errorf("send step text: %w", err)
	}

	return &State{Scenario: a.Name(), Step: actionIdx + 1}, nil
}

// Ensure ActionsScenario implements Scenario interface at compile time.
var _ Scenario = ActionsScenario{}

package scenario

import (
	"context"
	"fmt"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

// MentionScenario demonstrates group-unique features: user mentions and silent notifications.
type MentionScenario struct{}

// NewMentionScenario creates a new mention demo scenario.
func NewMentionScenario() *MentionScenario {
	return &MentionScenario{}
}

// Name returns the scenario command name.
func (MentionScenario) Name() string { return "mention" }

// Description returns a short Russian description for /help.
func (MentionScenario) Description() string {
	return "демо упоминаний и тихих сообщений"
}

// ChatType returns Group — this scenario only works in group chats.
func (MentionScenario) ChatType() ChatType { return ChatTypeGroup }

// RequiredPermissions returns nil — no admin permissions needed.
func (MentionScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the mention demo scenario.
func (ms MentionScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	switch s.Step {
	case 0:
		return ms.step0(ctx, p, s)
	case 1:
		return ms.step1(ctx, p, s)
	case 2:
		return ms.step2(p)
	default:
		return nil, fmt.Errorf("mention: unknown step %d", s.Step)
	}
}

// step0: send a formatted message using markdown to demonstrate group text formatting.
func (MentionScenario) step0(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID
	userID := upd.UserID

	text := "1. Форматирование сообщений (FormatMarkdown)\n\n" +
		"В групповых чатах можно использовать **жирный текст**, *курсив* и другие форматирования. " +
		"Это сообщение отправлено с `SetFormat(model.FormatMarkdown)`.\n\n" +
		"Нажмите «Далее»."

	msg := maxbotcli.NewMessage().
		SetChat(chatID).
		SetUser(userID).
		SetText(text).
		SetFormat(model.FormatMarkdown).
		AddKeyboard(navKeyboard())

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("send mention message: %w", err)
	}

	s.Step = 1

	return s, nil
}

// step1: send a silent notification using WithoutNotify.
func (MentionScenario) step1(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID
	userID := upd.UserID

	msg := maxbotcli.NewMessage().
		SetChat(chatID).
		SetUser(userID).
		SetText("2. Тихое сообщение (WithoutNotify)\n\n" +
			"Участники чата не получили уведомление об этом сообщении. " +
			"Это удобно для информационных сообщений, которые не требуют немедленного внимания.").
		WithoutNotify().
		AddKeyboard(navKeyboard())

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("send silent message: %w", err)
	}

	s.Step = 2

	return s, nil
}

// step2: scenario complete.
func (MentionScenario) step2(p Params) (*State, error) {
	if err := p.Ctx.Send("Демо упоминаний и тихих сообщений завершено!"); err != nil {
		return nil, fmt.Errorf("send completion: %w", err)
	}

	return nil, ErrDone
}

// Ensure MentionScenario implements Scenario interface at compile time.
var _ Scenario = MentionScenario{}

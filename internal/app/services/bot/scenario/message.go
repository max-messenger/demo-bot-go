package scenario

import (
	"context"
	"fmt"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
	"go.uber.org/zap"
)

const (
	messageIDKey = "message_id"
	messageName  = "message"
)

// MessageScenario demonstrates the Messages API: send, edit, reply, delete.
type MessageScenario struct{}

// Name returns the scenario command name.
func (MessageScenario) Name() string { return messageName }

// Description returns a short Russian description for /help.
func (MessageScenario) Description() string { return "демо работы с сообщениями" }

// ChatType returns Personal — this scenario only works in direct messages.
func (MessageScenario) ChatType() ChatType { return ChatTypePersonal }

// RequiredPermissions returns nil — no admin permissions needed for personal chats.
func (MessageScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the message demo scenario.
func (m MessageScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	switch s.Step {
	case 0:
		return m.step0(ctx, p, s)
	case 1:
		return m.step1(ctx, p, s)
	case 2:
		return m.step2(ctx, p, s)
	case 3:
		return m.step3(ctx, p, s)
	case 4:
		return m.step4(p)
	default:
		p.Log.Error("unknown step", zap.Int("step", s.Step))

		return nil, fmt.Errorf("unknown step: %d", s.Step)
	}
}

func (m MessageScenario) step0(ctx context.Context, p Params, s *State) (*State, error) {
	msg := maxbotcli.NewMessage().
		SetText("1. Отправка сообщения").
		SetUser(p.Ctx.Update().UserID).
		SetChat(p.Ctx.Update().ChatID).
		AddKeyboard(navKeyboard())

	result, err := p.API.Client().Messages.Send(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	if s.Data == nil {
		s.Data = make(map[string]any)
	}

	s.Data[messageIDKey] = result.Message.Body.Mid
	s.Step = 1

	if err := p.Ctx.Send("Сообщение отправлено с помощью Messages.Send(). Нажмите «Далее» для редактирования."); err != nil {
		return nil, fmt.Errorf("send explanation: %w", err)
	}

	return s, nil
}

func (m MessageScenario) step1(ctx context.Context, p Params, s *State) (*State, error) {
	messageID, err := getMessageID(s, messageIDKey)
	if err != nil {
		return nil, err
	}

	body := model.NewMessageBody{
		Text: "2. Сообщение отредактировано!",
		Attachments: []model.Attachment{
			navKeyboard().Build(),
		},
	}

	if _, err := p.API.Client().Messages.EditMessage(ctx, messageID, body); err != nil {
		return nil, fmt.Errorf("edit message: %w", err)
	}

	if err := p.Ctx.Send("Сообщение отредактировано с помощью Messages.EditMessage(). Нажмите «Далее» для ответа."); err != nil {
		return nil, fmt.Errorf("send explanation: %w", err)
	}

	s.Step = 2

	return s, nil
}

func (m MessageScenario) step2(ctx context.Context, p Params, s *State) (*State, error) {
	messageID, err := getMessageID(s, messageIDKey)
	if err != nil {
		return nil, err
	}

	msg := maxbotcli.NewMessage().
		SetReply("3. Ответ на сообщение", messageID).
		SetUser(p.Ctx.Update().UserID).
		SetChat(p.Ctx.Update().ChatID).
		AddKeyboard(navKeyboard())

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("reply to message: %w", err)
	}

	if err := p.Ctx.Send("Ответ отправлен с помощью Messages.Send() с Link типа reply. Нажмите «Далее» для удаления."); err != nil {
		return nil, fmt.Errorf("send explanation: %w", err)
	}

	s.Step = 3

	return s, nil
}

func (m MessageScenario) step3(ctx context.Context, p Params, s *State) (*State, error) {
	messageID, err := getMessageID(s, messageIDKey)
	if err != nil {
		return nil, err
	}

	if _, err := p.API.Client().Messages.DeleteMessage(ctx, messageID); err != nil {
		return nil, fmt.Errorf("delete message: %w", err)
	}

	if err := p.Ctx.Send("Сообщение удалено с помощью Messages.DeleteMessage().", maxbot.WithKeyboard(cancelKeyboard())); err != nil {
		return nil, fmt.Errorf("send final message: %w", err)
	}

	s.Step = 4

	return s, nil
}

func (m MessageScenario) step4(p Params) (*State, error) {
	if err := p.Ctx.Send("Демо сообщений завершено!"); err != nil {
		return nil, fmt.Errorf("send completion message: %w", err)
	}

	return nil, ErrDone
}

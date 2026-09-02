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
	chatMessageIDKey  = "chat_message_id"
	chatOriginalTitle = "chat_original_title"
)

// ChatScenario demonstrates group chat operations: pin, unpin, get pinned, edit title.
type ChatScenario struct{}

// NewChatScenario creates a new chat demo scenario.
func NewChatScenario() *ChatScenario {
	return &ChatScenario{}
}

// Name returns the scenario command name.
func (ChatScenario) Name() string { return "chat" }

// Description returns a short Russian description for /help.
func (ChatScenario) Description() string {
	return "демо работы с чатом: закрепление, открепление, изменение названия"
}

// ChatType returns Group — this scenario only works in group chats.
func (ChatScenario) ChatType() ChatType { return ChatTypeGroup }

// RequiredPermissions returns admin permissions needed for the chat scenario.
func (ChatScenario) RequiredPermissions() []string {
	return []string{"pin_message", "change_chat_info"}
}

// Handle processes one step of the chat demo scenario.
func (c ChatScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	switch s.Step {
	case 0:
		return c.step0(ctx, p, s)
	case 1:
		return c.step1(ctx, p, s)
	case 2:
		return c.step2(ctx, p, s)
	case 3:
		return c.step3(ctx, p, s)
	case 4:
		return c.step4(ctx, p, s)
	default:
		p.Log.Error("unknown step", zap.Int("step", s.Step))

		return nil, fmt.Errorf("chat: unknown step %d", s.Step)
	}
}

// step0: send a message, pin it, save messageID and original title.
func (c ChatScenario) step0(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	// Send a message and get its ID from the response.
	result, err := p.API.Client().Messages.Send(ctx, maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(chatID).
		SetText("Сообщение для закрепления"),
	)
	if err != nil {
		return nil, fmt.Errorf("send message to pin: %w", err)
	}

	messageID := result.Message.Body.Mid

	// Get the chat to save the original title.
	chat, err := p.API.Client().Chats.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat info: %w", err)
	}

	if _, err := p.API.Client().Chats.PinMessage(ctx, chatID, messageID, false); err != nil {
		return nil, fmt.Errorf("pin message: %w", err)
	}

	kb := navKeyboard()

	if err := p.Ctx.Send("Сообщение закреплено с помощью Chats.PinMessage(). Нажмите «Далее».", maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send pin confirmation: %w", err)
	}

	if s.Data == nil {
		s.Data = make(map[string]any)
	}

	s.Data[chatMessageIDKey] = messageID
	s.Data[chatOriginalTitle] = chat.Title
	s.Step = 1

	return s, nil
}

// step1: get pinned message, display info, show "Открепить" button.
func (c ChatScenario) step1(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	result, err := p.API.Client().Chats.GetPinnedMessage(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get pinned message: %w", err)
	}

	pinnedText := result.Message.Body.Text
	pinnedMid := result.Message.Body.Mid

	text := fmt.Sprintf("Закреплённое сообщение (Chats.GetPinnedMessage):\n\n"+
		"ID: %s\nТекст: %s", pinnedMid, pinnedText)

	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("Открепить", "next")
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	if err := p.Ctx.Send(text, maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send pinned info: %w", err)
	}

	s.Step = 2

	return s, nil
}

// step2: unpin message.
func (c ChatScenario) step2(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	if _, err := p.API.Client().Chats.UnpinMessage(ctx, chatID); err != nil {
		return nil, fmt.Errorf("unpin message: %w", err)
	}

	kb := navKeyboard()

	if err := p.Ctx.Send("Сообщение откреплено с помощью Chats.UnpinMessage(). Нажмите «Далее».", maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send unpin confirmation: %w", err)
	}

	s.Step = 3

	return s, nil
}

// step3: edit chat title (append " [demo]").
func (c ChatScenario) step3(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	originalTitle, err := getStateString(s, chatOriginalTitle)
	if err != nil {
		return nil, err
	}

	newTitle := originalTitle + " [demo]"

	if _, err := p.API.Client().Chats.EditChat(ctx, chatID, model.ChatPatch{
		Title: newTitle,
	}); err != nil {
		return nil, fmt.Errorf("edit chat title: %w", err)
	}

	kb := navKeyboard()

	if err := p.Ctx.Send(
		fmt.Sprintf("Название изменено на «%s» с помощью Chats.EditChat(). Нажмите «Далее» для восстановления.", newTitle),
		maxbot.WithKeyboard(kb),
	); err != nil {
		return nil, fmt.Errorf("send edit confirmation: %w", err)
	}

	s.Step = 4

	return s, nil
}

// step4: restore original chat title, scenario complete.
func (c ChatScenario) step4(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	originalTitle, err := getStateString(s, chatOriginalTitle)
	if err != nil {
		return nil, err
	}

	if _, err := p.API.Client().Chats.EditChat(ctx, chatID, model.ChatPatch{
		Title: originalTitle,
	}); err != nil {
		return nil, fmt.Errorf("restore chat title: %w", err)
	}

	if err := p.Ctx.Send("Демо чата завершено!"); err != nil {
		return nil, fmt.Errorf("send completion message: %w", err)
	}

	return nil, ErrDone
}

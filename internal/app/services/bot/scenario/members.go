package scenario

import (
	"context"
	"fmt"
	"strings"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
	"go.uber.org/zap"
)

const (
	membersSelectedUserIDKey = "members_selected_user_id"
	memberPayloadPrefix      = "member:"
)

// MembersScenario demonstrates group member operations: get membership, list members, remove, add.
type MembersScenario struct{}

// NewMembersScenario creates a new members demo scenario.
func NewMembersScenario() *MembersScenario {
	return &MembersScenario{}
}

// Name returns the scenario command name.
func (MembersScenario) Name() string { return "members" }

// Description returns a short Russian description for /help.
func (MembersScenario) Description() string {
	return "демо работы с участниками: удаление и добавление"
}

// ChatType returns Group — this scenario only works in group chats.
func (MembersScenario) ChatType() ChatType { return ChatTypeGroup }

// RequiredPermissions returns admin permissions needed for the members scenario.
func (MembersScenario) RequiredPermissions() []string {
	return []string{"add_remove_members"}
}

// Handle processes one step of the members demo scenario.
func (m MembersScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
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
		return m.step4(ctx, p, s)
	case 5:
		return m.step5(p)
	default:
		p.Log.Error("unknown step", zap.Int("step", s.Step))

		return nil, fmt.Errorf("members: unknown step %d", s.Step)
	}
}

// step0: get current membership info.
func (m MembersScenario) step0(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	membership, err := p.API.Client().Chats.GetMembership(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get membership: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Текущее участие в чате (Chats.GetMembership):\n\n")
	fmt.Fprintf(&sb, "Имя: %s\n", membership.Name)
	fmt.Fprintf(&sb, "UserID: %d\n", membership.UserID)
	fmt.Fprintf(&sb, "Владелец: %v\n", membership.IsOwner)
	fmt.Fprintf(&sb, "Администратор: %v\n", membership.IsAdmin)

	if len(membership.Permissions) > 0 {
		sb.WriteString("Права: ")
		for i, perm := range membership.Permissions {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(string(perm))
		}
		sb.WriteString("\n")
	}

	kb := navKeyboard()

	if err := p.Ctx.Send(sb.String(), maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send membership info: %w", err)
	}

	s.Step = 1

	return s, nil
}

// step1: get members list.
func (m MembersScenario) step1(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	members, err := p.API.Client().Chats.GetMembers(ctx, chatID, 0, 10, nil)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Участники чата (Chats.GetMembers, первые 10):\n\n")
	fmt.Fprintf(&sb, "Всего получено: %d\n\n", len(members.Members))

	for i, member := range members.Members {
		fmt.Fprintf(&sb, "%d. %s (ID: %d)\n", i+1, member.Name, member.UserID)
	}

	kb := navKeyboard()

	if err := p.Ctx.Send(sb.String(), maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send members list: %w", err)
	}

	s.Step = 2

	return s, nil
}

// step2: show member buttons for removal selection.
func (m MembersScenario) step2(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()

	// Check if user selected a member via callback.
	if upd.Callback != nil && strings.HasPrefix(upd.Callback.Payload, memberPayloadPrefix) {
		return m.handleMemberSelection(p, s)
	}

	chatID := upd.ChatID

	candidates, err := m.getRemovableMembers(ctx, p, chatID)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		if err := p.Ctx.Send("Нет участников для демонстрации удаления. " +
			"Демо участников завершено!"); err != nil {
			return nil, fmt.Errorf("send no candidates: %w", err)
		}

		return nil, ErrDone
	}

	if err := m.sendRemovableMemberSelection(p, candidates); err != nil {
		return nil, err
	}

	s.Step = 2

	return s, nil
}

// handleMemberSelection processes the callback when a member is selected for removal.
func (m MembersScenario) handleMemberSelection(p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()

	userIDStr := strings.TrimPrefix(upd.Callback.Payload, memberPayloadPrefix)
	userID, err := parseUserID(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("parse selected user id: %w", err)
	}

	if s.Data == nil {
		s.Data = make(map[string]any)
	}

	s.Data[membersSelectedUserIDKey] = userID
	s.Step = 3

	return s, nil
}

// getRemovableMembers returns members who are not bots and not owners (up to 5).
func (m MembersScenario) getRemovableMembers(ctx context.Context, p Params, chatID int64) ([]model.ChatMember, error) {
	members, err := p.API.Client().Chats.GetMembers(ctx, chatID, 0, 20, nil)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	var candidates []model.ChatMember

	for _, member := range members.Members {
		if !member.IsBot && !member.IsOwner {
			candidates = append(candidates, member)
		}

		if len(candidates) >= 5 {
			break
		}
	}

	return candidates, nil
}

// sendRemovableMemberSelection sends a keyboard with member buttons for removal.
func (m MembersScenario) sendRemovableMemberSelection(p Params, candidates []model.ChatMember) error {
	kb := model.NewKeyboard()

	for _, c := range candidates {
		kb.AddRow().AddCallBack(c.Name, fmt.Sprintf("%s%d", memberPayloadPrefix, c.UserID))
	}

	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	text := "Выберите участника для удаления из чата (Chats.RemoveMember):"

	if err := p.Ctx.Send(text, maxbot.WithKeyboard(kb)); err != nil {
		return fmt.Errorf("send member selection: %w", err)
	}

	return nil
}

// step3: remove the selected member.
func (m MembersScenario) step3(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	selectedUserID, err := getStateInt64(s, membersSelectedUserIDKey)
	if err != nil {
		return nil, err
	}

	remRes, err := p.API.Client().Chats.RemoveMember(ctx, chatID, selectedUserID, false)
	if err != nil {
		return nil, fmt.Errorf("remove member: %w", err)
	}

	if !remRes.Success {
		p.Log.Warn("remove member returned non-success",
			zap.String("message", remRes.Message),
			zap.Int64("user_id", selectedUserID),
		)
	}

	kb := navKeyboard()

	if err := p.Ctx.Send(
		fmt.Sprintf("Пользователь (ID: %d) удалён с помощью Chats.RemoveMember(). "+
			"Теперь добавим его обратно. Нажмите «Далее».", selectedUserID),
		maxbot.WithKeyboard(kb),
	); err != nil {
		return nil, fmt.Errorf("send remove confirmation: %w", err)
	}

	s.Step = 4

	return s, nil
}

// step4: add the removed member back.
func (m MembersScenario) step4(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	selectedUserID, err := getStateInt64(s, membersSelectedUserIDKey)
	if err != nil {
		return nil, err
	}

	// Warn about privacy requirement using markdown bold.
	warnMsg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(chatID).
		SetText("Важно: у пользователя в настройках приватности должно быть **«Пригласить в чат → могут все»**. " +
			"Иначе добавление не сработает.").
		SetFormat(model.FormatMarkdown)

	if _, sendErr := p.API.Client().Messages.Send(ctx, warnMsg); sendErr != nil {
		return nil, fmt.Errorf("send privacy warning: %w", sendErr)
	}

	res, err := p.API.Client().Chats.AddMembers(ctx, chatID, []int64{selectedUserID})
	if err != nil {
		return nil, fmt.Errorf("add member: %w", err)
	}

	if !res.Success {
		msg := fmt.Sprintf("Chats.AddMembers() вернул ошибку: %s", res.Message)
		if res.Message == "" {
			msg = "Chats.AddMembers() завершился неуспешно (success=false). " +
				"Возможно, пользователь запретил добавление в чаты в настройках приватности."
		}

		if err := p.Ctx.Send(msg); err != nil {
			return nil, fmt.Errorf("send add failure: %w", err)
		}

		return nil, ErrDone
	}

	kb := navKeyboard()

	if err := p.Ctx.Send(
		fmt.Sprintf("Пользователь (ID: %d) добавлен обратно с помощью Chats.AddMembers(). "+
			"Нажмите «Далее».", selectedUserID),
		maxbot.WithKeyboard(kb),
	); err != nil {
		return nil, fmt.Errorf("send add confirmation: %w", err)
	}

	s.Step = 5

	return s, nil
}

// step5: scenario complete.
func (m MembersScenario) step5(p Params) (*State, error) {
	if err := p.Ctx.Send("Демо участников завершено!"); err != nil {
		return nil, fmt.Errorf("send completion: %w", err)
	}

	return nil, ErrDone
}

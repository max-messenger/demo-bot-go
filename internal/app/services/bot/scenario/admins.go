package scenario

import (
	"context"
	"fmt"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
	"go.uber.org/zap"
)

const (
	adminsSelectedUserIDKey = "admins_selected_user_id"
	adminPayloadPrefix      = "admin:"
)

// AdminsScenario demonstrates group admin operations: get admins, set admin, delete admin.
type AdminsScenario struct{}

// NewAdminsScenario creates a new admins demo scenario.
func NewAdminsScenario() *AdminsScenario {
	return &AdminsScenario{}
}

// Name returns the scenario command name.
func (AdminsScenario) Name() string { return "admins" }

// Description returns a short Russian description for /help.
func (AdminsScenario) Description() string {
	return "демо работы с администраторами: назначение, просмотр, удаление"
}

// ChatType returns Group — this scenario only works in group chats.
func (AdminsScenario) ChatType() ChatType { return ChatTypeGroup }

// RequiredPermissions returns admin permissions needed for the admins scenario.
func (AdminsScenario) RequiredPermissions() []string {
	return []string{"add_admins"}
}

// Handle processes one step of the admins demo scenario.
func (a AdminsScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	switch s.Step {
	case 0:
		return a.step0(ctx, p, s)
	case 1:
		return a.step1(ctx, p, s)
	case 2:
		return a.step2(ctx, p, s)
	case 3:
		return a.step3(ctx, p, s)
	case 4:
		return a.step4(ctx, p, s)
	default:
		p.Log.Error("unknown step", zap.Int("step", s.Step))

		return nil, fmt.Errorf("admins: unknown step %d", s.Step)
	}
}

// step0: get current admins and display their permissions.
func (a AdminsScenario) step0(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	admins, err := p.API.Client().Chats.GetAdmins(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get admins: %w", err)
	}

	text := formatAdminsList("Текущие администраторы чата (Chats.GetAdmins)", admins.Members)

	kb := navKeyboard()

	if err := p.Ctx.Send(text, maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send admins info: %w", err)
	}

	s.Step = 1

	return s, nil
}

// step1: show non-admin members as buttons for selection.
func (a AdminsScenario) step1(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()

	// Check if user selected a member via callback.
	if upd.Callback != nil && strings.HasPrefix(upd.Callback.Payload, adminPayloadPrefix) {
		return a.handleAdminSelection(p, s)
	}

	chatID := upd.ChatID

	candidates, err := a.getNonAdminMembers(ctx, p, chatID)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		if err := p.Ctx.Send("Нет участников без прав администратора для демонстрации. " +
			"Демо администраторов завершено!"); err != nil {
			return nil, fmt.Errorf("send no candidates: %w", err)
		}

		return nil, ErrDone
	}

	if err := a.sendMemberSelection(p, candidates); err != nil {
		return nil, err
	}

	s.Step = 1

	return s, nil
}

// handleAdminSelection processes the callback when a member is selected for admin.
func (a AdminsScenario) handleAdminSelection(p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()

	userIDStr := strings.TrimPrefix(upd.Callback.Payload, adminPayloadPrefix)
	userID, err := parseUserID(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("parse selected user id: %w", err)
	}

	if s.Data == nil {
		s.Data = make(map[string]any)
	}

	s.Data[adminsSelectedUserIDKey] = userID
	s.Step = 2

	return s, nil
}

// getNonAdminMembers returns members who are not admins and not bots (up to 5).
func (a AdminsScenario) getNonAdminMembers(ctx context.Context, p Params, chatID int64) ([]model.ChatMember, error) {
	admins, err := p.API.Client().Chats.GetAdmins(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get admins for filter: %w", err)
	}

	adminIDs := make(map[int64]bool, len(admins.Members))
	for _, admin := range admins.Members {
		adminIDs[admin.UserID] = true
	}

	members, err := p.API.Client().Chats.GetMembers(ctx, chatID, 0, 20, nil)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	var candidates []model.ChatMember

	for _, member := range members.Members {
		if !adminIDs[member.UserID] && !member.IsBot {
			candidates = append(candidates, member)
		}

		if len(candidates) >= 5 {
			break
		}
	}

	return candidates, nil
}

// sendMemberSelection sends a keyboard with member buttons.
func (a AdminsScenario) sendMemberSelection(p Params, candidates []model.ChatMember) error {
	kb := model.NewKeyboard()

	for _, c := range candidates {
		kb.AddRow().AddCallBack(c.Name, fmt.Sprintf("%s%d", adminPayloadPrefix, c.UserID))
	}

	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	text := "Выберите участника для назначения администратором (Chats.SetAdmins):"

	if err := p.Ctx.Send(text, maxbot.WithKeyboard(kb)); err != nil {
		return fmt.Errorf("send member selection: %w", err)
	}

	return nil
}

// step2: set admin — send only the new admin (SetAdmins is additive, not replacement).
func (a AdminsScenario) step2(ctx context.Context, p Params, s *State) (*State, error) {
	selectedUserID, err := getStateInt64(s, adminsSelectedUserIDKey)
	if err != nil {
		return nil, err
	}

	upd := p.Ctx.Update()
	chatID := upd.ChatID

	// Grant the same permissions the bot has, so the API doesn't reject the request.
	botMembership, err := p.API.Client().Chats.GetMembership(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get bot membership: %w", err)
	}

	res, err := p.API.Client().Chats.SetAdmins(ctx, chatID, []model.ChatAdmin{
		{UserID: selectedUserID, Permissions: botMembership.Permissions},
	})
	if err != nil {
		return nil, fmt.Errorf("set admin: %w", err)
	}

	if !res.Success {
		msg := fmt.Sprintf("Chats.SetAdmins() вернул ошибку: %s", res.Message)
		if res.Message == "" {
			msg = "Chats.SetAdmins() завершился неуспешно (success=false)"
		}

		if err := p.Ctx.Send(msg); err != nil {
			return nil, fmt.Errorf("send set admin failure: %w", err)
		}

		return nil, ErrDone
	}

	kb := navKeyboard()

	if err := p.Ctx.Send(
		fmt.Sprintf(
			"Пользователь (ID: %d) назначен администратором с помощью Chats.SetAdmins(). "+
				"Нажмите «Далее».", selectedUserID,
		),
		maxbot.WithKeyboard(kb),
	); err != nil {
		return nil, fmt.Errorf("send set admin confirmation: %w", err)
	}

	s.Step = 3

	return s, nil
}

// step3: get admins again and show updated list.
func (a AdminsScenario) step3(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	admins, err := p.API.Client().Chats.GetAdmins(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get admins after set: %w", err)
	}

	text := formatAdminsList("Обновлённый список администраторов (Chats.GetAdmins)", admins.Members)

	kb := navKeyboard()

	if err := p.Ctx.Send(text, maxbot.WithKeyboard(kb)); err != nil {
		return nil, fmt.Errorf("send updated admins info: %w", err)
	}

	s.Step = 4

	return s, nil
}

// formatAdminsList formats an admin list into a readable string.
func formatAdminsList(header string, members []model.ChatMember) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s:\n\n", header)

	if len(members) == 0 {
		sb.WriteString("Администраторы не найдены.\n")
	}

	for i, admin := range members {
		fmt.Fprintf(&sb, "%d. %s (ID: %d)\n", i+1, admin.Name, admin.UserID)

		if len(admin.Permissions) > 0 {
			sb.WriteString("   Права: ")

			for j, perm := range admin.Permissions {
				if j > 0 {
					sb.WriteString(", ")
				}

				sb.WriteString(string(perm))
			}

			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// step4: remove admin rights using DeleteAdmins, scenario complete.
func (a AdminsScenario) step4(ctx context.Context, p Params, s *State) (*State, error) {
	upd := p.Ctx.Update()
	chatID := upd.ChatID

	selectedUserID, err := getStateInt64(s, adminsSelectedUserIDKey)
	if err != nil {
		return nil, err
	}

	if _, err := p.API.Client().Chats.DeleteAdmins(ctx, chatID, selectedUserID); err != nil {
		return nil, fmt.Errorf("delete admin: %w", err)
	}

	confirm := fmt.Sprintf(
		"Права администратора сняты с пользователя (ID: %d) с помощью Chats.DeleteAdmins(). "+
			"Демо администраторов завершено!", selectedUserID,
	)

	if err := p.Ctx.Send(confirm); err != nil {
		return nil, fmt.Errorf("send delete admin confirmation: %w", err)
	}

	return nil, ErrDone
}

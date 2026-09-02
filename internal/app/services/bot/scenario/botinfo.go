package scenario

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
)

// BotInfoScenario demonstrates Bots.GetMyInfo and Bots.EditMyInfo API methods.
type BotInfoScenario struct{}

// NewBotInfoScenario creates a new botinfo scenario.
func NewBotInfoScenario() *BotInfoScenario {
	return &BotInfoScenario{}
}

// Name returns the scenario command name.
func (BotInfoScenario) Name() string { return "botinfo" }

// Description returns a short Russian description for /help.
func (BotInfoScenario) Description() string {
	return "информация о боте и её редактирование"
}

// ChatType returns where this scenario can run.
func (BotInfoScenario) ChatType() ChatType { return ChatTypePersonal }

// RequiredPermissions returns admin permissions needed (group only).
func (BotInfoScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the botinfo scenario.
func (s BotInfoScenario) Handle(ctx context.Context, p Params, st *State) (*State, error) {
	switch st.Step {
	case 0:
		return s.handleStep0(ctx, p, st)
	case 1:
		return s.handleStep1(ctx, p, st)
	case 2:
		return s.handleStep2(ctx, p, st)
	default:
		return nil, fmt.Errorf("botinfo: unknown step %d", st.Step)
	}
}

func (s BotInfoScenario) handleStep0(ctx context.Context, p Params, st *State) (*State, error) {
	info, err := p.API.Client().Bots.GetMyInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("get my info: %w", err)
	}

	origDesc := info.Description
	if st.Data == nil {
		st.Data = make(map[string]any)
	}

	st.Data["original_description"] = origDesc

	text := formatBotInfo(info)

	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("Изменить", "bi:edit")
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	if sendErr := p.Ctx.Send(text, maxbot.WithKeyboard(kb)); sendErr != nil {
		return nil, fmt.Errorf("send bot info: %w", sendErr)
	}

	st.Step = 1

	return st, nil
}

func (s BotInfoScenario) handleStep1(ctx context.Context, p Params, st *State) (*State, error) {
	callback := p.Ctx.Update().Callback
	if callback == nil {
		if err := p.Ctx.Send("Нажмите кнопку «Изменить» или «Завершить» ниже."); err != nil {
			return nil, fmt.Errorf("send hint: %w", err)
		}

		return st, nil
	}

	switch callback.Payload {
	case "bi:edit":
		return s.editBotInfo(ctx, p, st)
	case PayloadCancel:
		return nil, ErrDone
	default:
		return st, nil
	}
}

func (s BotInfoScenario) editBotInfo(ctx context.Context, p Params, st *State) (*State, error) {
	info, err := p.API.Client().Bots.EditMyInfo(ctx, model.BotPatch{
		Description: fmt.Sprintf("Описание изменено демо-ботом %s", time.Now().Format("2006-01-02 15:04:05")),
	})
	if err != nil {
		return nil, fmt.Errorf("edit my info: %w", err)
	}

	text := formatBotInfo(info)

	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("Восстановить", "bi:restore")
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	if sendErr := p.Ctx.Send(text, maxbot.WithKeyboard(kb)); sendErr != nil {
		return nil, fmt.Errorf("send updated bot info: %w", sendErr)
	}

	st.Step = 2

	return st, nil
}

func (s BotInfoScenario) handleStep2(ctx context.Context, p Params, st *State) (*State, error) {
	callback := p.Ctx.Update().Callback
	if callback == nil {
		if err := p.Ctx.Send("Нажмите кнопку «Восстановить» или «Завершить» ниже."); err != nil {
			return nil, fmt.Errorf("send hint: %w", err)
		}

		return st, nil
	}

	switch callback.Payload {
	case "bi:restore":
		return s.restoreBotInfo(ctx, p, st)
	case PayloadCancel:
		return nil, ErrDone
	default:
		return st, nil
	}
}

func (s BotInfoScenario) restoreBotInfo(ctx context.Context, p Params, st *State) (*State, error) {
	origDesc, ok := st.Data["original_description"].(string)
	if !ok {
		return nil, fmt.Errorf("original_description not found in state data")
	}

	_, err := p.API.Client().Bots.EditMyInfo(ctx, model.BotPatch{
		Description: origDesc,
	})
	if err != nil {
		return nil, fmt.Errorf("restore bot info: %w", err)
	}

	if sendErr := p.Ctx.Send("Демо информации о боте завершено!"); sendErr != nil {
		return nil, fmt.Errorf("send completion message: %w", sendErr)
	}

	return nil, ErrDone
}

func formatBotInfo(info model.BotInfo) string {
	var sb strings.Builder

	sb.WriteString("Информация о боте:\n\n")
	fmt.Fprintf(&sb, "Имя: %s\n", info.FirstName)
	fmt.Fprintf(&sb, "Описание: %s\n", info.Description)

	if len(info.Commands) > 0 {
		sb.WriteString("Команды:\n")
		for _, cmd := range info.Commands {
			fmt.Fprintf(&sb, "  /%s", cmd.Name)
			if cmd.Description != "" {
				fmt.Fprintf(&sb, " — %s", cmd.Description)
			}
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// Ensure BotInfoScenario implements Scenario interface at compile time.
var _ Scenario = BotInfoScenario{}

package scenario_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"demo_bot/internal/app/services/bot/mocks"
	"demo_bot/internal/app/services/bot/scenario"
	scenariomocks "demo_bot/internal/app/services/bot/scenario/mocks"
)

func TestUnitSendStartMenu_Personal(t *testing.T) {
	t.Parallel()

	p, ctrl := setupStartTest(t)
	defer ctrl.Finish()

	scenarios := []scenario.Scenario{
		&stubScenario{name: "message", description: "демо сообщений"},
		&stubScenario{name: "attachments", description: "демо вложений"},
	}

	p.ctxMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		DoAndReturn(func(text string, _ ...any) error {
			assert.Contains(t, text, "Личные сценарии")
			assert.Contains(t, text, "/message")
			assert.Contains(t, text, "/attachments")

			return nil
		})

	err := scenario.SendStartMenu(p.params, scenarios, true)

	require.NoError(t, err)
}

func TestUnitSendStartMenu_Group(t *testing.T) {
	t.Parallel()

	p, ctrl := setupStartTest(t)
	defer ctrl.Finish()

	scenarios := []scenario.Scenario{
		&stubScenario{name: "mention", description: "демо упоминаний"},
	}

	p.ctxMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		DoAndReturn(func(text string, _ ...any) error {
			assert.Contains(t, text, "Групповые сценарии")
			assert.Contains(t, text, "/mention")

			return nil
		})

	err := scenario.SendStartMenu(p.params, scenarios, false)

	require.NoError(t, err)
}

func TestUnitSendStartMenu_SendError(t *testing.T) {
	t.Parallel()

	p, ctrl := setupStartTest(t)
	defer ctrl.Finish()

	p.ctxMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(errors.New("send failed"))

	err := scenario.SendStartMenu(p.params, []scenario.Scenario{}, true)

	require.Error(t, err)
}

func TestUnitSendHelpText_Success(t *testing.T) {
	t.Parallel()

	p, ctrl := setupStartTest(t)
	defer ctrl.Finish()

	scenarios := []scenario.Scenario{
		&stubScenario{name: "message", description: "демо сообщений"},
		&stubScenario{name: "leave", description: "демо выхода из чата"},
	}

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, msg *maxbotcli.Message) (model.SendMessageResult, error) {
			text := msgText(t, msg)

			assert.Contains(t, text, "/message")
			assert.Contains(t, text, "/leave")
			assert.Contains(t, text, "/start")

			return model.SendMessageResult{}, nil
		})

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	err := scenario.SendHelpText(p.params, scenarios)

	require.NoError(t, err)
}

func TestUnitSendHelpText_WithPermissions(t *testing.T) {
	t.Parallel()

	p, ctrl := setupStartTest(t)
	defer ctrl.Finish()

	scenarios := []scenario.Scenario{
		&stubScenario{
			name:        "admins",
			description: "демо администрирования",
			permissions: []string{"chat:write", "members:read"},
		},
	}

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, msg *maxbotcli.Message) (model.SendMessageResult, error) {
			text := msgText(t, msg)

			assert.Contains(t, text, "/admins")
			assert.Contains(t, text, "Требуемые права")
			assert.Contains(t, text, "**chat:write**")
			assert.Contains(t, text, "**members:read**")

			return model.SendMessageResult{}, nil
		})

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	err := scenario.SendHelpText(p.params, scenarios)

	require.NoError(t, err)
}

func TestUnitSendHelpText_SendError(t *testing.T) {
	t.Parallel()

	p, ctrl := setupStartTest(t)
	defer ctrl.Finish()

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("send failed"))

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	err := scenario.SendHelpText(p.params, []scenario.Scenario{})

	require.Error(t, err)
}

// stubScenario is a minimal Scenario implementation used in start/help tests.
type stubScenario struct {
	name        string
	description string
	permissions []string
}

func (s *stubScenario) Name() string                  { return s.name }
func (s *stubScenario) Description() string           { return s.description }
func (s *stubScenario) ChatType() scenario.ChatType   { return scenario.ChatTypeAny }
func (s *stubScenario) RequiredPermissions() []string { return s.permissions }

func (s *stubScenario) Handle(
	_ context.Context,
	_ scenario.Params,
	st *scenario.State,
) (*scenario.State, error) {
	return st, nil
}

// startTestParams holds mocks and params for start/help tests.
type startTestParams struct {
	apiMock *scenariomocks.MockAPI
	ctxMock *mocks.MockContext
	params  scenario.Params
}

// setupStartTest creates mocks and params for start/help tests.
func setupStartTest(t *testing.T) (startTestParams, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	apiMock := scenariomocks.NewMockAPI(ctrl)
	ctxMock := mocks.NewMockContext(ctrl)

	ctxMock.EXPECT().Update().Return(model.Update{
		ChatID: testChatID,
		UserID: testUserID,
	}).AnyTimes()

	ctxMock.EXPECT().Context().Return(context.Background()).AnyTimes()

	params := scenario.Params{
		API: apiMock,
		Ctx: ctxMock,
		Log: zap.NewNop(),
	}

	return startTestParams{
		apiMock: apiMock,
		ctxMock: ctxMock,
		params:  params,
	}, ctrl
}

// msgText extracts the text from a maxbotcli.Message for assertion purposes.
func msgText(t *testing.T, msg *maxbotcli.Message) string {
	t.Helper()

	body := msg.MessageBody()

	return strings.TrimSpace(body.Text)
}

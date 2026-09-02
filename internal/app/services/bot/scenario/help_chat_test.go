package scenario_test

import (
	"context"
	"errors"
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

func TestUnitSendGroupHelpText_Success(t *testing.T) {
	t.Parallel()

	p, ctrl := setupHelpChatTest(t)
	defer ctrl.Finish()

	scenarios := []scenario.Scenario{
		&stubScenario{name: "mention", description: "демо упоминаний"},
		&stubScenario{name: "leave", description: "демо выхода из чата"},
	}

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, msg *maxbotcli.Message) (model.SendMessageResult, error) {
			text := msgText(t, msg)

			assert.Contains(t, text, "/mention")
			assert.Contains(t, text, "/leave")
			assert.Contains(t, text, "/start")
			assert.Contains(t, text, "/help")

			return model.SendMessageResult{}, nil
		})

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	err := scenario.SendGroupHelpText(p.params, scenarios)

	require.NoError(t, err)
}

func TestUnitSendGroupHelpText_WithPermissions(t *testing.T) {
	t.Parallel()

	p, ctrl := setupHelpChatTest(t)
	defer ctrl.Finish()

	scenarios := []scenario.Scenario{
		&stubScenario{
			name:        "admins",
			description: "демо администрирования",
			permissions: []string{"chat:write"},
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

			return model.SendMessageResult{}, nil
		})

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	err := scenario.SendGroupHelpText(p.params, scenarios)

	require.NoError(t, err)
}

func TestUnitSendGroupHelpText_SendError(t *testing.T) {
	t.Parallel()

	p, ctrl := setupHelpChatTest(t)
	defer ctrl.Finish()

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("send failed"))

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	err := scenario.SendGroupHelpText(p.params, []scenario.Scenario{})

	require.Error(t, err)
}

// helpChatTestParams holds mocks and params for help_chat tests.
type helpChatTestParams struct {
	apiMock *scenariomocks.MockAPI
	ctxMock *mocks.MockContext
	params  scenario.Params
}

// setupHelpChatTest creates mocks and params for help_chat tests.
func setupHelpChatTest(t *testing.T) (helpChatTestParams, *gomock.Controller) {
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

	return helpChatTestParams{
		apiMock: apiMock,
		ctxMock: ctxMock,
		params:  params,
	}, ctrl
}

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
	scenariomocks "demo_bot/internal/app/services/bot/scenario/mocks"
	"demo_bot/internal/app/services/bot/scenario"
)

func TestUnitMentionScenario_Step0(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupMentionTest(t)
	defer ctrl.Finish()

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	state, err := svc.Handle(context.Background(), p.params, &scenario.State{Step: 0})

	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, 1, state.Step)
}

func TestUnitMentionScenario_Step1(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupMentionTest(t)
	defer ctrl.Finish()

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	state, err := svc.Handle(context.Background(), p.params, &scenario.State{Step: 1})

	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, 2, state.Step)
}

func TestUnitMentionScenario_Step2(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupMentionTest(t)
	defer ctrl.Finish()

	p.ctxMock.EXPECT().
		Send("Демо упоминаний и тихих сообщений завершено!").
		Return(nil)

	state, err := svc.Handle(context.Background(), p.params, &scenario.State{Step: 2})

	require.ErrorIs(t, err, scenario.ErrDone)
	assert.Nil(t, state)
}

func TestUnitMentionScenario_UnknownStep(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupMentionTest(t)
	defer ctrl.Finish()

	state, err := svc.Handle(context.Background(), p.params, &scenario.State{Step: 99})

	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "unknown step 99")
}

func TestUnitMentionScenario_Step0_SendError(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupMentionTest(t)
	defer ctrl.Finish()

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("network error"))

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	state, err := svc.Handle(context.Background(), p.params, &scenario.State{Step: 0})

	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "send mention message")
}

func TestUnitMentionScenario_Step1_SendError(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupMentionTest(t)
	defer ctrl.Finish()

	msgsMock := mocks.NewMockMessagesAPI(ctrl)
	msgsMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("network error"))

	apiClient := &maxbotcli.Api{Messages: msgsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	state, err := svc.Handle(context.Background(), p.params, &scenario.State{Step: 1})

	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "send silent message")
}

func TestUnitMentionScenario_Step2_SendError(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupMentionTest(t)
	defer ctrl.Finish()

	p.ctxMock.EXPECT().
		Send("Демо упоминаний и тихих сообщений завершено!").
		Return(errors.New("send failed"))

	state, err := svc.Handle(context.Background(), p.params, &scenario.State{Step: 2})

	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "send completion")
}

func TestUnitMentionScenario_Metadata(t *testing.T) {
	t.Parallel()

	svc := scenario.NewMentionScenario()

	assert.Equal(t, "mention", svc.Name())
	assert.Equal(t, "демо упоминаний и тихих сообщений", svc.Description())
	assert.Equal(t, scenario.ChatTypeGroup, svc.ChatType())
	assert.Nil(t, svc.RequiredPermissions())
}

// mentionTestParams holds mocks and params for mention scenario tests.
type mentionTestParams struct {
	apiMock *scenariomocks.MockAPI
	ctxMock *mocks.MockContext
	params  scenario.Params
}

// setupMentionTest creates mocks and params for mention scenario tests.
func setupMentionTest(t *testing.T) (*scenario.MentionScenario, mentionTestParams, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	apiMock := scenariomocks.NewMockAPI(ctrl)
	ctxMock := mocks.NewMockContext(ctrl)

	ctxMock.EXPECT().Update().Return(model.Update{
		ChatID: testChatID,
		UserID: testUserID,
	}).AnyTimes()

	params := scenario.Params{
		API: apiMock,
		Ctx: ctxMock,
		Log: zap.NewNop(),
	}

	return scenario.NewMentionScenario(), mentionTestParams{
		apiMock: apiMock,
		ctxMock: ctxMock,
		params:  params,
	}, ctrl
}

package scenario_test

import (
	"context"
	"errors"
	"fmt"
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

func TestUnitLeaveScenario_Confirmation(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupLeaveTest(t)
	defer ctrl.Finish()

	p.expectUpdateNoCallback(t)
	mockSendWithKeyboard(t, p.ctxMock)

	state, err := svc.Handle(context.Background(), p.params, nil)

	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, "leave", state.Scenario)
	assert.Equal(t, 0, state.Step)
}

func TestUnitLeaveScenario_LeaveYes(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupLeaveTest(t)
	defer ctrl.Finish()

	p.expectUpdateWithCallback(t, "leave:yes")

	chatsMock := mocks.NewMockChatsAPI(ctrl)
	chatsMock.EXPECT().
		LeaveChat(gomock.Any(), testChatID).
		Return(model.SimpleQueryResult{Success: true}, nil)

	apiClient := &maxbotcli.Api{Chats: chatsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	state, err := svc.Handle(context.Background(), p.params, nil)

	require.ErrorIs(t, err, scenario.ErrDoneNoMenu)
	assert.Nil(t, state)
}

func TestUnitLeaveScenario_LeaveNo(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupLeaveTest(t)
	defer ctrl.Finish()

	p.expectUpdateWithCallback(t, "leave:no")
	p.ctxMock.EXPECT().Send("Отменено").Return(nil)

	state, err := svc.Handle(context.Background(), p.params, nil)

	require.ErrorIs(t, err, scenario.ErrDone)
	assert.Nil(t, state)
}

func TestUnitLeaveScenario_UnknownCallback(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupLeaveTest(t)
	defer ctrl.Finish()

	p.expectUpdateWithCallback(t, "unknown")
	mockSendWithKeyboard(t, p.ctxMock)

	state, err := svc.Handle(context.Background(), p.params, nil)

	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, "leave", state.Scenario)
	assert.Equal(t, 0, state.Step)
}

func TestUnitLeaveScenario_LeaveChatError(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupLeaveTest(t)
	defer ctrl.Finish()

	p.expectUpdateWithCallback(t, "leave:yes")

	chatsMock := mocks.NewMockChatsAPI(ctrl)
	chatsMock.EXPECT().
		LeaveChat(gomock.Any(), testChatID).
		Return(model.SimpleQueryResult{}, errors.New("network error"))

	apiClient := &maxbotcli.Api{Chats: chatsMock}
	p.apiMock.EXPECT().Client().Return(apiClient)

	state, err := svc.Handle(context.Background(), p.params, nil)

	assert.Nil(t, state)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "leave chat")
}

func TestUnitLeaveScenario_SendCancelError(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupLeaveTest(t)
	defer ctrl.Finish()

	p.expectUpdateWithCallback(t, "leave:no")
	p.ctxMock.EXPECT().Send("Отменено").Return(errors.New("send failed"))

	state, err := svc.Handle(context.Background(), p.params, nil)

	assert.Nil(t, state)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send cancel message")
}

func TestUnitLeaveScenario_SendConfirmationError(t *testing.T) {
	t.Parallel()

	svc, p, ctrl := setupLeaveTest(t)
	defer ctrl.Finish()

	p.expectUpdateNoCallback(t)
	p.ctxMock.EXPECT().
		Send("Покинуть чат?", gomock.Any()).
		Return(errors.New("send failed"))

	state, err := svc.Handle(context.Background(), p.params, nil)

	assert.Nil(t, state)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send leave confirmation")
}

func TestUnitLeaveScenario_Metadata(t *testing.T) {
	t.Parallel()

	svc := scenario.NewLeaveScenario()

	assert.Equal(t, "leave", svc.Name())
	assert.Equal(t, "демо выхода из чата", svc.Description())
	assert.Equal(t, scenario.ChatTypeGroup, svc.ChatType())
	assert.Nil(t, svc.RequiredPermissions())
}

const (
	testChatID int64 = 100
	testUserID int64 = 200
)

// leaveTestParams holds mocks and params for leave scenario tests.
type leaveTestParams struct {
	apiMock *scenariomocks.MockAPI
	ctxMock *mocks.MockContext
	params  scenario.Params
}

// setupLeaveTest creates mocks and params for leave scenario tests.
func setupLeaveTest(t *testing.T) (*scenario.LeaveScenario, leaveTestParams, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	apiMock := scenariomocks.NewMockAPI(ctrl)
	ctxMock := mocks.NewMockContext(ctrl)

	params := scenario.Params{
		API: apiMock,
		Ctx: ctxMock,
		Log: zap.NewNop(),
	}

	return scenario.NewLeaveScenario(), leaveTestParams{
		apiMock: apiMock,
		ctxMock: ctxMock,
		params:  params,
	}, ctrl
}

// expectUpdateNoCallback sets up an Update expectation with no callback.
func (p *leaveTestParams) expectUpdateNoCallback(t *testing.T) {
	t.Helper()

	p.ctxMock.EXPECT().Update().Return(model.Update{
		ChatID: testChatID,
		UserID: testUserID,
	}).AnyTimes()
}

// expectUpdateWithCallback sets up an Update expectation with a callback payload.
func (p *leaveTestParams) expectUpdateWithCallback(t *testing.T, payload string) {
	t.Helper()

	p.ctxMock.EXPECT().Update().Return(model.Update{
		ChatID: testChatID,
		UserID: testUserID,
		Callback: &model.Callback{
			Payload:    payload,
			CallbackID: fmt.Sprintf("cb-%s", payload),
		},
	}).AnyTimes()
}

// mockSendWithKeyboard sets up a Send expectation for keyboard confirmation.
func mockSendWithKeyboard(t *testing.T, ctxMock *mocks.MockContext) {
	t.Helper()

	ctxMock.EXPECT().
		Send("Покинуть чат?", gomock.Any()).
		Return(nil)
}

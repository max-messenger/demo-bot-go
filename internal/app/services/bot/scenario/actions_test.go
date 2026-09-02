package scenario_test

import (
	"context"
	"errors"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"demo_bot/internal/app/services/bot/mocks"
	scenariomocks "demo_bot/internal/app/services/bot/scenario/mocks"
	"demo_bot/internal/app/services/bot/scenario"
)

type ActionsScenarioSuite struct {
	suite.Suite

	ctrl    *gomock.Controller
	api     *scenariomocks.MockAPI
	ctx     *mocks.MockContext
	chats   *mocks.MockChatsAPI
	messages *mocks.MockMessagesAPI
	log     *zap.Logger
}

func (s *ActionsScenarioSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.api = scenariomocks.NewMockAPI(s.ctrl)
	s.ctx = mocks.NewMockContext(s.ctrl)
	s.chats = mocks.NewMockChatsAPI(s.ctrl)
	s.messages = mocks.NewMockMessagesAPI(s.ctrl)
	s.log = zap.NewNop()

	apiClient := &maxbot.Api{
		Chats:    s.chats,
		Messages: s.messages,
	}
	s.api.EXPECT().Client().Return(apiClient).AnyTimes()
}

func (s *ActionsScenarioSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ActionsScenarioSuite) params() scenario.Params {
	s.T().Helper()
	return scenario.Params{
		API: s.api,
		Ctx: s.ctx,
		Log: s.log,
	}
}

func TestUnitActionsScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ActionsScenarioSuite))
}

func (s *ActionsScenarioSuite) TestStep0_Intro() {
	s.ctx.EXPECT().Update().Return(model.Update{ChatID: 42}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), "actions", state.Scenario)
	assert.Equal(s.T(), 0, state.Step)
}

func (s *ActionsScenarioSuite) TestStep0_Cancel() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "cancel"},
	}).AnyTimes()

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 0},
	)

	require.ErrorIs(s.T(), err, scenario.ErrDone)
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestStep0_StartCallback() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:next"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionTypingOn).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), "actions", state.Scenario)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *ActionsScenarioSuite) TestStep1_Next() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:next"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionSendingPhoto).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 2, state.Step)
}

func (s *ActionsScenarioSuite) TestStep1_Repeat() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:repeat"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionTypingOn).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *ActionsScenarioSuite) TestStep1_Cancel() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "cancel"},
	}).AnyTimes()

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.ErrorIs(s.T(), err, scenario.ErrDone)
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestStep1_TextHint() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		Message: &model.MessageUpdate{Body: model.MessageBody{Text: "hello"}},
	}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *ActionsScenarioSuite) TestLastStep_ReturnsErrDone() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:next"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionMarkSeen).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 5},
	)

	require.ErrorIs(s.T(), err, scenario.ErrDone)
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestSendActionError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:next"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionTypingOn).
		Return(model.SimpleQueryResult{}, errors.New("network error"))

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestStep0_UnknownCallback() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "unknown"},
	}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), "actions", state.Scenario)
	assert.Equal(s.T(), 0, state.Step)
}

func (s *ActionsScenarioSuite) TestStep0_IntroSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{ChatID: 42}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send intro")
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestResolveActionStep_UnknownCallback() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "unknown_payload"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionTypingOn).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *ActionsScenarioSuite) TestStep1_TextHintSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:  42,
		Message: &model.MessageUpdate{Body: model.MessageBody{Text: "hello"}},
	}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("hint error"))

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send hint")
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestStep1_NoCallbackNoMessage() {
	s.ctx.EXPECT().Update().Return(model.Update{ChatID: 42}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionTypingOn).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *ActionsScenarioSuite) TestSendStepResult_LastStepSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:next"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), model.ActionMarkSeen).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 5},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send final message")
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestSendStepResult_NonLastStepSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:next"},
	}).AnyTimes()
	s.chats.EXPECT().SendAction(gomock.Any(), int64(42), gomock.Any()).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send step text")
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestHandleAction_UnknownStep() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "act:next"},
	}).AnyTimes()

	sc := scenario.NewActionsScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "actions", Step: 100},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "actions: unknown step")
	assert.Nil(s.T(), state)
}

func (s *ActionsScenarioSuite) TestMetadata() {
	sc := scenario.NewActionsScenario()

	assert.Equal(s.T(), "actions", sc.Name())
	assert.Equal(s.T(), "демо индикаторов действий в чате", sc.Description())
	assert.Equal(s.T(), scenario.ChatTypePersonal, sc.ChatType())
	assert.Nil(s.T(), sc.RequiredPermissions())
}

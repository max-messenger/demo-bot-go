package scenario_test

import (
	"context"
	"fmt"
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

type KeyboardScenarioSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	api      *scenariomocks.MockAPI
	ctx      *mocks.MockContext
	chats    *mocks.MockChatsAPI
	messages *mocks.MockMessagesAPI
	log      *zap.Logger
}

func (s *KeyboardScenarioSuite) SetupTest() {
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

func (s *KeyboardScenarioSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *KeyboardScenarioSuite) params() scenario.Params {
	s.T().Helper()
	return scenario.Params{
		API: s.api,
		Ctx: s.ctx,
		Log: s.log,
	}
}

func TestUnitKeyboardScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(KeyboardScenarioSuite))
}

func (s *KeyboardScenarioSuite) TestStep0_IntroNoCallback() {
	s.ctx.EXPECT().Update().Return(model.Update{ChatID: 42}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), "keyboard", state.Scenario)
	assert.Equal(s.T(), 0, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep0_CallbackAnswer() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		Callback: &model.Callback{
			Payload:    "kb:callback",
			CallbackID: "cb123",
		},
	}).AnyTimes()
	s.messages.EXPECT().AnswerOnCallback(
		gomock.Any(), "cb123", gomock.Any(),
	).Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep0_Cancel() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID:   42,
		Callback: &model.Callback{Payload: "cancel"},
	}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	// cancel at step 0 returns step 0 (unknown callback falls through to intro)
	assert.Equal(s.T(), 0, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep1_LinkButton() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 2, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep2_GeoLocation() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 2},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 3, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep3_Contact() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 3},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 4, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep4_Clipboard() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 4},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 5, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep5_MessageButton() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 5},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 6, state.Step)
}

func (s *KeyboardScenarioSuite) TestStep6_Done() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 6},
	)

	require.ErrorIs(s.T(), err, scenario.ErrDone)
	assert.Nil(s.T(), state)
}

func (s *KeyboardScenarioSuite) TestDefaultStep_ErrDone() {
	sc := scenario.KeyboardScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 99},
	)

	require.ErrorIs(s.T(), err, scenario.ErrDone)
	assert.Nil(s.T(), state)
}

func (s *KeyboardScenarioSuite) TestMetadata() {
	sc := scenario.KeyboardScenario{}

	assert.Equal(s.T(), "keyboard", sc.Name())
	assert.Equal(s.T(), "демо клавиатур и кнопок", sc.Description())
	assert.Equal(s.T(), scenario.ChatTypePersonal, sc.ChatType())
	assert.Nil(s.T(), sc.RequiredPermissions())
}

func (s *KeyboardScenarioSuite) TestStep0_CallbackAnswerError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		Callback: &model.Callback{
			Payload:    "kb:callback",
			CallbackID: "cb123",
		},
	}).AnyTimes()
	s.messages.EXPECT().AnswerOnCallback(
		gomock.Any(), "cb123", gomock.Any(),
	).Return(model.SimpleQueryResult{}, fmt.Errorf("api error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "answer on callback")
}

func (s *KeyboardScenarioSuite) TestStep0_CallbackExplanationSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		Callback: &model.Callback{
			Payload:    "kb:callback",
			CallbackID: "cb123",
		},
	}).AnyTimes()
	s.messages.EXPECT().AnswerOnCallback(
		gomock.Any(), "cb123", gomock.Any(),
	).Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send explanation")
}

func (s *KeyboardScenarioSuite) TestStep0_IntroSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{ChatID: 42}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send callback demo")
}

func (s *KeyboardScenarioSuite) TestStep1_LinkSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send link demo")
}

func (s *KeyboardScenarioSuite) TestStep2_GeoKeyboardSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 2},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send geo keyboard")
}

func (s *KeyboardScenarioSuite) TestStep2_NavSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 2},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send nav")
}

func (s *KeyboardScenarioSuite) TestStep3_ContactKeyboardSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 3},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send contact keyboard")
}

func (s *KeyboardScenarioSuite) TestStep3_NavSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 3},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send nav")
}

func (s *KeyboardScenarioSuite) TestStep4_ClipboardSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 4},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send clipboard demo")
}

func (s *KeyboardScenarioSuite) TestStep5_MessageButtonSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 5},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send message button keyboard")
}

func (s *KeyboardScenarioSuite) TestStep5_NavSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 5},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send nav")
}

func (s *KeyboardScenarioSuite) TestStep6_FinalSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(fmt.Errorf("send error"))

	sc := scenario.KeyboardScenario{}
	_, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "keyboard", Step: 6},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send final message")
}

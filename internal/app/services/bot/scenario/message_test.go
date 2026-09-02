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

type MessageScenarioSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	api      *scenariomocks.MockAPI
	ctx      *mocks.MockContext
	chats    *mocks.MockChatsAPI
	messages *mocks.MockMessagesAPI
	log      *zap.Logger
}

func (s *MessageScenarioSuite) SetupTest() {
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

func (s *MessageScenarioSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *MessageScenarioSuite) params() scenario.Params {
	s.T().Helper()

	return scenario.Params{
		API: s.api,
		Ctx: s.ctx,
		Log: s.log,
	}
}

func TestUnitMessageScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MessageScenarioSuite))
}

func (s *MessageScenarioSuite) TestStep0_Send() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{
			Message: model.Message{
				Body: model.MessageBody{Mid: "msg42"},
			},
		}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "message", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
	assert.Equal(s.T(), "msg42", state.Data["message_id"])
}

func (s *MessageScenarioSuite) TestStep0_SendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("send failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "message", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep1_Edit() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().EditMessage(gomock.Any(), "msg42", gomock.Any()).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     1,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 2, state.Step)
}

func (s *MessageScenarioSuite) TestStep2_Reply() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     2,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 3, state.Step)
}

func (s *MessageScenarioSuite) TestStep3_Delete() {
	s.messages.EXPECT().DeleteMessage(gomock.Any(), "msg42").
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     3,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 4, state.Step)
}

func (s *MessageScenarioSuite) TestStep4_Done() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "message", Step: 4},
	)

	require.ErrorIs(s.T(), err, scenario.ErrDone)
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestUnknownStep_Error() {
	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "message", Step: 99},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep0_ExplanationSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{
			Message: model.Message{
				Body: model.MessageBody{Mid: "msg42"},
			},
		}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("explanation failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "message", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send explanation")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep1_MissingMessageID() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     1,
			Data:     map[string]any{},
		},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep1_EditMessageError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().EditMessage(gomock.Any(), "msg42", gomock.Any()).
		Return(model.SimpleQueryResult{}, errors.New("edit failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     1,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "edit message")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep1_ExplanationSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().EditMessage(gomock.Any(), "msg42", gomock.Any()).
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("explanation failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     1,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send explanation")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep2_MissingMessageID() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     2,
			Data:     nil,
		},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep2_ReplyError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("reply failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     2,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "reply to message")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep2_ExplanationSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
	}).AnyTimes()
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("explanation failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     2,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send explanation")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep3_MissingMessageID() {
	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     3,
			Data:     map[string]any{},
		},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep3_DeleteMessageError() {
	s.messages.EXPECT().DeleteMessage(gomock.Any(), "msg42").
		Return(model.SimpleQueryResult{}, errors.New("delete failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     3,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "delete message")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep3_FinalSendError() {
	s.messages.EXPECT().DeleteMessage(gomock.Any(), "msg42").
		Return(model.SimpleQueryResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("send failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{
			Scenario: "message",
			Step:     3,
			Data:     map[string]any{"message_id": "msg42"},
		},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send final message")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestStep4_CompletionSendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("send failed"))

	sc := scenario.MessageScenario{}
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "message", Step: 4},
	)

	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "send completion message")
	assert.Nil(s.T(), state)
}

func (s *MessageScenarioSuite) TestMetadata() {
	sc := scenario.MessageScenario{}

	assert.Equal(s.T(), "message", sc.Name())
	assert.Equal(s.T(), "демо работы с сообщениями", sc.Description())
	assert.Equal(s.T(), scenario.ChatTypePersonal, sc.ChatType())
	assert.Nil(s.T(), sc.RequiredPermissions())
}

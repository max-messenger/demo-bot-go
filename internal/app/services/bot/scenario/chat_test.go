package scenario_test

import (
	"context"
	"errors"
	"testing"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"demo_bot/internal/app/services/bot/mocks"
	"demo_bot/internal/app/services/bot/scenario"
	scenariomocks "demo_bot/internal/app/services/bot/scenario/mocks"
)

func TestUnitChatScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(chatSuite))
}

type chatSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	api   *scenariomocks.MockAPI
	ctx   *mocks.MockContext
	msgs  *mocks.MockMessagesAPI
	chats *mocks.MockChatsAPI
	log   *zap.Logger
	s     *scenario.ChatScenario
}

func (su *chatSuite) SetupTest() {
	su.ctrl = gomock.NewController(su.T())
	su.api = scenariomocks.NewMockAPI(su.ctrl)
	su.ctx = mocks.NewMockContext(su.ctrl)
	su.msgs = mocks.NewMockMessagesAPI(su.ctrl)
	su.chats = mocks.NewMockChatsAPI(su.ctrl)
	su.log = zap.NewNop()
	su.s = scenario.NewChatScenario()

	apiClient := &maxbotcli.Api{
		Messages: su.msgs,
		Chats:    su.chats,
	}
	su.api.EXPECT().Client().Return(apiClient).AnyTimes()
}

func (su *chatSuite) TearDownTest() {
	su.ctrl.Finish()
}

// params builds scenario.Params for the chat scenario.
func (su *chatSuite) params() scenario.Params {
	su.T().Helper()

	return scenario.Params{
		API: su.api,
		Ctx: su.ctx,
		Log: su.log,
	}
}

// chatGroupUpdate returns a group update for chat scenario tests.
func chatGroupUpdate() model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 789,
		Message: &model.MessageUpdate{
			Recipient: model.Recipient{ChatType: model.ChatTypeChat},
			Body:      model.MessageBody{Text: ""},
		},
	}
}

func (su *chatSuite) TestStep0_SendPinMessage() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	sendResult := model.SendMessageResult{
		Message: model.Message{
			Body: model.MessageBody{Mid: "msg-123"},
		},
	}
	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(sendResult, nil)
	su.chats.EXPECT().GetChat(gomock.Any(), int64(789)).Return(model.Chat{Title: "Test Chat"}, nil)
	su.chats.EXPECT().PinMessage(gomock.Any(), int64(789), "msg-123", false).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "chat",
		Step:     0,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
	assert.Equal(su.T(), "msg-123", state.Data["chat_message_id"])
	assert.Equal(su.T(), "Test Chat", state.Data["chat_original_title"])
}

func (su *chatSuite) TestStep0_SendMessageError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "chat",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send message to pin")
}

func (su *chatSuite) TestStep0_GetChatError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	sendResult := model.SendMessageResult{
		Message: model.Message{Body: model.MessageBody{Mid: "msg-1"}},
	}
	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(sendResult, nil)
	su.chats.EXPECT().GetChat(gomock.Any(), int64(789)).Return(model.Chat{}, errors.New("get chat error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "chat",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get chat info")
}

func (su *chatSuite) TestStep0_PinMessageError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	sendResult := model.SendMessageResult{
		Message: model.Message{Body: model.MessageBody{Mid: "msg-1"}},
	}
	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(sendResult, nil)
	su.chats.EXPECT().GetChat(gomock.Any(), int64(789)).Return(model.Chat{Title: "Chat"}, nil)
	su.chats.EXPECT().PinMessage(gomock.Any(), int64(789), "msg-1", false).
		Return(model.SimpleQueryResult{}, errors.New("pin error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "chat",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "pin message")
}

func (su *chatSuite) TestStep1_GetPinnedMessage() {
	upd := chatGroupUpdate()
	upd.Callback = &model.Callback{Payload: "next"}
	upd.Message = nil
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	pinnedResult := model.GetPinnedMessageResult{
		Message: model.Message{
			Body: model.MessageBody{Mid: "msg-1", Text: "pinned text"},
		},
	}
	su.chats.EXPECT().GetPinnedMessage(gomock.Any(), int64(789)).Return(pinnedResult, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "chat",
		Step:     1,
		Data:     map[string]any{"chat_message_id": "msg-1", "chat_original_title": "Chat"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *chatSuite) TestStep2_UnpinMessage() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().UnpinMessage(gomock.Any(), int64(789)).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "chat",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 3, state.Step)
}

func (su *chatSuite) TestStep3_EditChatTitle() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().EditChat(gomock.Any(), int64(789), gomock.Any()).
		Return(model.Chat{}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "chat",
		Step:     3,
		Data:     map[string]any{"chat_original_title": "Original Title"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 4, state.Step)
}

func (su *chatSuite) TestStep4_RestoreTitle() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().EditChat(gomock.Any(), int64(789), gomock.Any()).
		Return(model.Chat{}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "chat",
		Step:     4,
		Data:     map[string]any{"chat_original_title": "Original Title"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *chatSuite) TestStep0_SendConfirmationError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	sendResult := model.SendMessageResult{
		Message: model.Message{Body: model.MessageBody{Mid: "msg-1"}},
	}
	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(sendResult, nil)
	su.chats.EXPECT().GetChat(gomock.Any(), int64(789)).Return(model.Chat{Title: "Chat"}, nil)
	su.chats.EXPECT().PinMessage(gomock.Any(), int64(789), "msg-1", false).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "chat",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send pin confirmation")
}

func (su *chatSuite) TestStep1_GetPinnedMessageError() {
	upd := chatGroupUpdate()
	upd.Callback = &model.Callback{Payload: "next"}
	upd.Message = nil
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetPinnedMessage(gomock.Any(), int64(789)).
		Return(model.GetPinnedMessageResult{}, errors.New("pinned error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     1,
		Data:     map[string]any{"chat_message_id": "msg-1", "chat_original_title": "Chat"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get pinned message")
}

func (su *chatSuite) TestStep1_SendError() {
	upd := chatGroupUpdate()
	upd.Callback = &model.Callback{Payload: "next"}
	upd.Message = nil
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	pinnedResult := model.GetPinnedMessageResult{
		Message: model.Message{
			Body: model.MessageBody{Mid: "msg-1", Text: "pinned text"},
		},
	}
	su.chats.EXPECT().GetPinnedMessage(gomock.Any(), int64(789)).Return(pinnedResult, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     1,
		Data:     map[string]any{"chat_message_id": "msg-1", "chat_original_title": "Chat"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send pinned info")
}

func (su *chatSuite) TestStep2_UnpinError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().UnpinMessage(gomock.Any(), int64(789)).
		Return(model.SimpleQueryResult{}, errors.New("unpin error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "unpin message")
}

func (su *chatSuite) TestStep2_SendError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().UnpinMessage(gomock.Any(), int64(789)).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send unpin confirmation")
}

func (su *chatSuite) TestStep3_GetStateStringError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "chat",
		Step:     3,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "chat_original_title not found")
}

func (su *chatSuite) TestStep3_EditChatError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().EditChat(gomock.Any(), int64(789), gomock.Any()).
		Return(model.Chat{}, errors.New("edit error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     3,
		Data:     map[string]any{"chat_original_title": "Original Title"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "edit chat title")
}

func (su *chatSuite) TestStep3_SendError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().EditChat(gomock.Any(), int64(789), gomock.Any()).
		Return(model.Chat{}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     3,
		Data:     map[string]any{"chat_original_title": "Original Title"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send edit confirmation")
}

func (su *chatSuite) TestStep4_GetStateStringError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "chat",
		Step:     4,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "chat_original_title not found")
}

func (su *chatSuite) TestStep4_EditChatError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().EditChat(gomock.Any(), int64(789), gomock.Any()).
		Return(model.Chat{}, errors.New("edit error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     4,
		Data:     map[string]any{"chat_original_title": "Original Title"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "restore chat title")
}

func (su *chatSuite) TestStep4_SendError() {
	upd := chatGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().EditChat(gomock.Any(), int64(789), gomock.Any()).
		Return(model.Chat{}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "chat",
		Step:     4,
		Data:     map[string]any{"chat_original_title": "Original Title"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send completion message")
}

func (su *chatSuite) TestUnknownStep() {
	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "chat",
		Step:     99,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "unknown step")
}

func TestUnitChatScenario_Metadata(t *testing.T) {
	t.Parallel()

	s := scenario.NewChatScenario()

	assert.Equal(t, "chat", s.Name())
	assert.Equal(t, scenario.ChatTypeGroup, s.ChatType())
	assert.Equal(t, []string{"pin_message", "change_chat_info"}, s.RequiredPermissions())
	assert.NotEmpty(t, s.Description())
}

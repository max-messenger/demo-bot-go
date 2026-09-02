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

func TestUnitMembersScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(membersSuite))
}

type membersSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	api   *scenariomocks.MockAPI
	ctx   *mocks.MockContext
	msgs  *mocks.MockMessagesAPI
	chats *mocks.MockChatsAPI
	log   *zap.Logger
	s     *scenario.MembersScenario
}

func (su *membersSuite) SetupTest() {
	su.ctrl = gomock.NewController(su.T())
	su.api = scenariomocks.NewMockAPI(su.ctrl)
	su.ctx = mocks.NewMockContext(su.ctrl)
	su.msgs = mocks.NewMockMessagesAPI(su.ctrl)
	su.chats = mocks.NewMockChatsAPI(su.ctrl)
	su.log = zap.NewNop()
	su.s = scenario.NewMembersScenario()

	apiClient := &maxbotcli.Api{
		Messages: su.msgs,
		Chats:    su.chats,
	}
	su.api.EXPECT().Client().Return(apiClient).AnyTimes()
}

func (su *membersSuite) TearDownTest() {
	su.ctrl.Finish()
}

// params builds scenario.Params for the members scenario.
func (su *membersSuite) params() scenario.Params {
	su.T().Helper()

	return scenario.Params{
		API: su.api,
		Ctx: su.ctx,
		Log: su.log,
	}
}

// membersGroupUpdate returns a group update for members scenario tests.
func membersGroupUpdate() model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 789,
		Message: &model.MessageUpdate{
			Recipient: model.Recipient{ChatType: model.ChatTypeChat},
			Body:      model.MessageBody{Text: ""},
		},
	}
}

func (su *membersSuite) TestStep0_GetMembership() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	membership := model.ChatMember{
		Name:        "TestBot",
		UserID:      100,
		IsOwner:     false,
		IsAdmin:     true,
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(membership, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "members",
		Step:     0,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
}

func (su *membersSuite) TestStep0_GetMembershipError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).
		Return(model.ChatMember{}, errors.New("membership error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "members",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get membership")
}

func (su *membersSuite) TestStep0_SendError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	membership := model.ChatMember{Name: "Bot", UserID: 100}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(membership, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "members",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send membership info")
}

func (su *membersSuite) TestStep1_GetMembers() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	membersList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Alice", UserID: 1},
			{Name: "Bob", UserID: 2},
		},
	}
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(10), gomock.Nil()).
		Return(membersList, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{Scenario: "members", Step: 1}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *membersSuite) TestStep1_GetMembersError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(10), gomock.Nil()).
		Return(model.ChatMembersList{}, errors.New("get members error"))

	st := &scenario.State{Scenario: "members", Step: 1}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get members")
}

func (su *membersSuite) TestStep5_Complete() {
	su.ctx.EXPECT().Send("Демо участников завершено!").Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "members",
		Step:     5,
	})

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *membersSuite) TestStep5_CompleteSendError() {
	su.ctx.EXPECT().Send("Демо участников завершено!").Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "members",
		Step:     5,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send completion")
}

func (su *membersSuite) TestUnknownStep() {
	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "members",
		Step:     99,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "unknown step")
}

// membersCallbackUpdate returns a group update with a callback payload.
func membersCallbackUpdate(payload string) model.Update {
	return model.Update{
		UserID:   123,
		ChatID:   789,
		Callback: &model.Callback{Payload: payload},
	}
}

// ---- Step 2 tests ----

func (su *membersSuite) TestStep2_CallbackSelection() {
	upd := membersCallbackUpdate("member:42")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "members",
		Step:     2,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 3, state.Step)
	assert.Equal(su.T(), int64(42), state.Data["members_selected_user_id"])
}

func (su *membersSuite) TestStep2_CallbackInvalidUserID() {
	upd := membersCallbackUpdate("member:abc")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "members",
		Step:     2,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "parse selected user id")
}

func (su *membersSuite) TestStep2_CallbackNilData() {
	upd := membersCallbackUpdate("member:55")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "members",
		Step:     2,
		Data:     nil,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 3, state.Step)
	assert.Equal(su.T(), int64(55), state.Data["members_selected_user_id"])
}

func (su *membersSuite) TestStep2_GetRemovableMembers() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	membersList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Alice", UserID: 1},
			{Name: "Bob", UserID: 2, IsBot: true},
			{Name: "Carol", UserID: 3, IsOwner: true},
			{Name: "Dave", UserID: 4},
		},
	}
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(membersList, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{Scenario: "members", Step: 2}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *membersSuite) TestStep2_GetMembersError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(model.ChatMembersList{}, errors.New("get members error"))

	st := &scenario.State{Scenario: "members", Step: 2}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get members")
}

func (su *membersSuite) TestStep2_NoCandidates() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	membersList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Bot", UserID: 1, IsBot: true},
			{Name: "Owner", UserID: 2, IsOwner: true},
		},
	}
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(membersList, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{Scenario: "members", Step: 2}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *membersSuite) TestStep2_NoCandidatesSendError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	membersList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Bot", UserID: 1, IsBot: true},
		},
	}
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(membersList, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{Scenario: "members", Step: 2}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send no candidates")
}

func (su *membersSuite) TestStep2_SendMemberSelectionError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	membersList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Alice", UserID: 1},
		},
	}
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(membersList, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{Scenario: "members", Step: 2}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send member selection")
}

// ---- Step 3 tests ----

func (su *membersSuite) TestStep3_RemoveMember() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().RemoveMember(gomock.Any(), int64(789), int64(42), false).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "members",
		Step:     3,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 4, state.Step)
}

func (su *membersSuite) TestStep3_NoSelectedUserID() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "members",
		Step:     3,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "members_selected_user_id not found")
}

func (su *membersSuite) TestStep3_RemoveMemberError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().RemoveMember(gomock.Any(), int64(789), int64(42), false).
		Return(model.SimpleQueryResult{}, errors.New("remove error"))

	st := &scenario.State{
		Scenario: "members",
		Step:     3,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "remove member")
}

func (su *membersSuite) TestStep3_RemoveMemberNotSuccess() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().RemoveMember(gomock.Any(), int64(789), int64(42), false).
		Return(model.SimpleQueryResult{Success: false, Message: "forbidden"}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "members",
		Step:     3,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 4, state.Step)
}

func (su *membersSuite) TestStep3_SendRemoveConfirmationError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().RemoveMember(gomock.Any(), int64(789), int64(42), false).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "members",
		Step:     3,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send remove confirmation")
}

// ---- Step 4 tests ----

func (su *membersSuite) TestStep4_AddMember() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.chats.EXPECT().AddMembers(gomock.Any(), int64(789), []int64{42}).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 5, state.Step)
}

func (su *membersSuite) TestStep4_NoSelectedUserID() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "members_selected_user_id not found")
}

func (su *membersSuite) TestStep4_SendPrivacyWarningError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("send warning error"))

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send privacy warning")
}

func (su *membersSuite) TestStep4_AddMemberError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.chats.EXPECT().AddMembers(gomock.Any(), int64(789), []int64{42}).
		Return(model.SimpleQueryResult{}, errors.New("add member error"))

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "add member")
}

func (su *membersSuite) TestStep4_AddMemberNotSuccessWithMessage() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.chats.EXPECT().AddMembers(gomock.Any(), int64(789), []int64{42}).
		Return(model.SimpleQueryResult{Success: false, Message: "forbidden"}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *membersSuite) TestStep4_AddMemberNotSuccessNoMessage() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.chats.EXPECT().AddMembers(gomock.Any(), int64(789), []int64{42}).
		Return(model.SimpleQueryResult{Success: false}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *membersSuite) TestStep4_AddMemberNotSuccessSendError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.chats.EXPECT().AddMembers(gomock.Any(), int64(789), []int64{42}).
		Return(model.SimpleQueryResult{Success: false, Message: "forbidden"}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send add failure")
}

func (su *membersSuite) TestStep4_SendAddConfirmationError() {
	upd := membersGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.chats.EXPECT().AddMembers(gomock.Any(), int64(789), []int64{42}).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "members",
		Step:     4,
		Data:     map[string]any{"members_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send add confirmation")
}

func TestUnitMembersScenario_Metadata(t *testing.T) {
	t.Parallel()

	s := scenario.NewMembersScenario()

	assert.Equal(t, "members", s.Name())
	assert.Equal(t, scenario.ChatTypeGroup, s.ChatType())
	assert.Equal(t, []string{"add_remove_members"}, s.RequiredPermissions())
	assert.NotEmpty(t, s.Description())
}

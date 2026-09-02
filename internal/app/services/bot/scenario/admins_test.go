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

func TestUnitAdminsScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(adminsSuite))
}

type adminsSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	api   *scenariomocks.MockAPI
	ctx   *mocks.MockContext
	msgs  *mocks.MockMessagesAPI
	chats *mocks.MockChatsAPI
	log   *zap.Logger
	s     *scenario.AdminsScenario
}

func (su *adminsSuite) SetupTest() {
	su.ctrl = gomock.NewController(su.T())
	su.api = scenariomocks.NewMockAPI(su.ctrl)
	su.ctx = mocks.NewMockContext(su.ctrl)
	su.msgs = mocks.NewMockMessagesAPI(su.ctrl)
	su.chats = mocks.NewMockChatsAPI(su.ctrl)
	su.log = zap.NewNop()
	su.s = scenario.NewAdminsScenario()

	apiClient := &maxbotcli.Api{
		Messages: su.msgs,
		Chats:    su.chats,
	}
	su.api.EXPECT().Client().Return(apiClient).AnyTimes()
}

func (su *adminsSuite) TearDownTest() {
	su.ctrl.Finish()
}

// params builds scenario.Params for the admins scenario.
func (su *adminsSuite) params() scenario.Params {
	su.T().Helper()

	return scenario.Params{
		API: su.api,
		Ctx: su.ctx,
		Log: su.log,
	}
}

// adminsGroupUpdate returns a group update for admins scenario tests.
func adminsGroupUpdate() model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 789,
		Message: &model.MessageUpdate{
			Recipient: model.Recipient{ChatType: model.ChatTypeChat},
			Body:      model.MessageBody{Text: ""},
		},
	}
}

func (su *adminsSuite) TestStep0_GetAdmins() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	adminsList := model.ChatMembersList{
		Members: []model.ChatMember{
			{
				Name:   "Admin1",
				UserID: 10,
				Permissions: []model.ChatAdminPermission{
					"pin_message",
					"change_chat_info",
				},
			},
			{Name: "Admin2", UserID: 20},
		},
	}
	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).Return(adminsList, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "admins",
		Step:     0,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
}

func (su *adminsSuite) TestStep0_GetAdminsError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{}, errors.New("get admins error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "admins",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get admins")
}

func (su *adminsSuite) TestStep0_EmptyAdmins() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{Members: []model.ChatMember{}}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "admins",
		Step:     0,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
}

func (su *adminsSuite) TestStep0_SendError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{Members: []model.ChatMember{}}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "admins",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send admins info")
}

func (su *adminsSuite) TestStep4_DeleteAdmins() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().DeleteAdmins(gomock.Any(), int64(789), int64(42)).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     4,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *adminsSuite) TestStep4_DeleteAdminsError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().DeleteAdmins(gomock.Any(), int64(789), int64(42)).
		Return(model.SimpleQueryResult{}, errors.New("delete error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     4,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "delete admin")
}

func (su *adminsSuite) TestStep4_SendConfirmError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().DeleteAdmins(gomock.Any(), int64(789), int64(42)).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     4,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send delete admin confirmation")
}

func (su *adminsSuite) TestUnknownStep() {
	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "admins",
		Step:     99,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "unknown step")
}

// adminsCallbackUpdate returns a group update with a callback payload.
func adminsCallbackUpdate(payload string) model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 789,
		Callback: &model.Callback{Payload: payload},
	}
}

func (su *adminsSuite) TestStep1_CallbackSelection() {
	upd := adminsCallbackUpdate("admin:42")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
	assert.Equal(su.T(), int64(42), state.Data["admins_selected_user_id"])
}

func (su *adminsSuite) TestStep1_CallbackInvalidUserID() {
	upd := adminsCallbackUpdate("admin:abc")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "parse selected user id")
}

func (su *adminsSuite) TestStep1_CallbackNilData() {
	upd := adminsCallbackUpdate("admin:55")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
		Data:     nil,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
	assert.Equal(su.T(), int64(55), state.Data["admins_selected_user_id"])
}

func (su *adminsSuite) TestStep1_GetNonAdminMembers() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	adminsList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Admin1", UserID: 10},
		},
	}
	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).Return(adminsList, nil)

	membersList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Admin1", UserID: 10},
			{Name: "User1", UserID: 20},
			{Name: "User2", UserID: 30, IsBot: true},
			{Name: "User3", UserID: 40},
		},
	}
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(membersList, nil)

	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
}

func (su *adminsSuite) TestStep1_GetNonAdminMembersNoCandidates() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	adminsList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Admin1", UserID: 10},
		},
	}
	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).Return(adminsList, nil)

	membersList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Admin1", UserID: 10},
		},
	}
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(membersList, nil)

	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *adminsSuite) TestStep1_GetAdminsError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{}, errors.New("get admins error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get admins for filter")
}

func (su *adminsSuite) TestStep1_GetMembersError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{Members: []model.ChatMember{}}, nil)
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(model.ChatMembersList{}, errors.New("get members error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get members")
}

func (su *adminsSuite) TestStep1_SendMemberSelectionError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{Members: []model.ChatMember{}}, nil)
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(model.ChatMembersList{
			Members: []model.ChatMember{{Name: "User1", UserID: 20}},
		}, nil)

	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send member selection")
}

func (su *adminsSuite) TestStep1_SendNoCandidatesError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{Members: []model.ChatMember{}}, nil)
	su.chats.EXPECT().GetMembers(gomock.Any(), int64(789), int64(0), int64(20), gomock.Nil()).
		Return(model.ChatMembersList{Members: []model.ChatMember{}}, nil)

	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     1,
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send no candidates")
}

func (su *adminsSuite) TestStep2_SetAdmins() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botMembership := model.ChatMember{
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(botMembership, nil)
	su.chats.EXPECT().SetAdmins(gomock.Any(), int64(789), gomock.Any()).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 3, state.Step)
}

func (su *adminsSuite) TestStep2_NoSelectedUserID() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "admins_selected_user_id not found")
}

func (su *adminsSuite) TestStep2_GetMembershipError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).
		Return(model.ChatMember{}, errors.New("membership error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get bot membership")
}

func (su *adminsSuite) TestStep2_SetAdminsError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botMembership := model.ChatMember{
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(botMembership, nil)
	su.chats.EXPECT().SetAdmins(gomock.Any(), int64(789), gomock.Any()).
		Return(model.SimpleQueryResult{}, errors.New("set admins error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "set admin")
}

func (su *adminsSuite) TestStep2_SetAdminsNotSuccess() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botMembership := model.ChatMember{
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(botMembership, nil)
	su.chats.EXPECT().SetAdmins(gomock.Any(), int64(789), gomock.Any()).
		Return(model.SimpleQueryResult{Success: false, Message: "forbidden"}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *adminsSuite) TestStep2_SetAdminsNotSuccessNoMessage() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botMembership := model.ChatMember{
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(botMembership, nil)
	su.chats.EXPECT().SetAdmins(gomock.Any(), int64(789), gomock.Any()).
		Return(model.SimpleQueryResult{Success: false}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *adminsSuite) TestStep2_SetAdminsNotSuccessSendError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botMembership := model.ChatMember{
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(botMembership, nil)
	su.chats.EXPECT().SetAdmins(gomock.Any(), int64(789), gomock.Any()).
		Return(model.SimpleQueryResult{Success: false, Message: "forbidden"}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send set admin failure")
}

func (su *adminsSuite) TestStep2_SendConfirmError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botMembership := model.ChatMember{
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}
	su.chats.EXPECT().GetMembership(gomock.Any(), int64(789)).Return(botMembership, nil)
	su.chats.EXPECT().SetAdmins(gomock.Any(), int64(789), gomock.Any()).
		Return(model.SimpleQueryResult{Success: true}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     2,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send set admin confirmation")
}

func (su *adminsSuite) TestStep3_GetAdminsAfterSet() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	adminsList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Admin1", UserID: 10, Permissions: []model.ChatAdminPermission{"pin_message"}},
			{Name: "NewAdmin", UserID: 42, Permissions: []model.ChatAdminPermission{}},
		},
	}
	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).Return(adminsList, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     3,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 4, state.Step)
}

func (su *adminsSuite) TestStep3_GetAdminsError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).
		Return(model.ChatMembersList{}, errors.New("get admins error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     3,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get admins after set")
}

func (su *adminsSuite) TestStep3_SendError() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	adminsList := model.ChatMembersList{
		Members: []model.ChatMember{
			{Name: "Admin1", UserID: 10},
		},
	}
	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).Return(adminsList, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "admins",
		Step:     3,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send updated admins info")
}

func (su *adminsSuite) TestStep3_EmptyAdminsList() {
	upd := adminsGroupUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	adminsList := model.ChatMembersList{Members: []model.ChatMember{}}
	su.chats.EXPECT().GetAdmins(gomock.Any(), int64(789)).Return(adminsList, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "admins",
		Step:     3,
		Data:     map[string]any{"admins_selected_user_id": int64(42)},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 4, state.Step)
}

func TestUnitAdminsScenario_Metadata(t *testing.T) {
	t.Parallel()

	s := scenario.NewAdminsScenario()

	assert.Equal(t, "admins", s.Name())
	assert.Equal(t, scenario.ChatTypeGroup, s.ChatType())
	assert.Equal(t, []string{"add_admins"}, s.RequiredPermissions())
	assert.NotEmpty(t, s.Description())
}

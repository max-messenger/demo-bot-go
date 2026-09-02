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

func TestUnitBotInfoScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(botInfoSuite))
}

type botInfoSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	api   *scenariomocks.MockAPI
	ctx   *mocks.MockContext
	msgs  *mocks.MockMessagesAPI
	bots  *mocks.MockBotsAPI
	chats *mocks.MockChatsAPI
	log   *zap.Logger
	s     scenario.BotInfoScenario
}

func (su *botInfoSuite) SetupTest() {
	su.ctrl = gomock.NewController(su.T())
	su.api = scenariomocks.NewMockAPI(su.ctrl)
	su.ctx = mocks.NewMockContext(su.ctrl)
	su.msgs = mocks.NewMockMessagesAPI(su.ctrl)
	su.bots = mocks.NewMockBotsAPI(su.ctrl)
	su.chats = mocks.NewMockChatsAPI(su.ctrl)
	su.log = zap.NewNop()

	apiClient := &maxbotcli.Api{
		Messages: su.msgs,
		Chats:    su.chats,
		Bots:     su.bots,
	}
	su.api.EXPECT().Client().Return(apiClient).AnyTimes()
}

func (su *botInfoSuite) TearDownTest() {
	su.ctrl.Finish()
}

// params builds scenario.Params for the botinfo scenario.
func (su *botInfoSuite) params() scenario.Params {
	su.T().Helper()

	return scenario.Params{
		API: su.api,
		Ctx: su.ctx,
		Log: su.log,
	}
}

func (su *botInfoSuite) TestStep0_GetMyInfo() {
	upd := botInfoPersonalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botInfo := model.BotInfo{
		FirstName:   "TestBot",
		Description: "Test description",
		Commands: []model.BotCommand{
			{Name: "start", Description: "Start bot"},
		},
	}
	su.bots.EXPECT().GetMyInfo(gomock.Any()).Return(botInfo, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "botinfo",
		Step:     0,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
	assert.Equal(su.T(), "Test description", state.Data["original_description"])
}

func (su *botInfoSuite) TestStep0_GetMyInfoError() {
	upd := botInfoPersonalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.bots.EXPECT().GetMyInfo(gomock.Any()).Return(model.BotInfo{}, errors.New("api error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "botinfo",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "get my info")
}

func (su *botInfoSuite) TestStep0_SendError() {
	upd := botInfoPersonalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	botInfo := model.BotInfo{FirstName: "Bot", Description: "Desc"}
	su.bots.EXPECT().GetMyInfo(gomock.Any()).Return(botInfo, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "botinfo",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send bot info")
}

func (su *botInfoSuite) TestStep1_CancelCallback() {
	upd := botInfoCallbackUpdate(scenario.PayloadCancel)
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "botinfo",
		Step:     1,
	})

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *botInfoSuite) TestStep1_EditCallback() {
	upd := botInfoCallbackUpdate("bi:edit")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	editedInfo := model.BotInfo{
		FirstName:   "TestBot",
		Description: "Описание изменено демо-ботом",
	}
	su.bots.EXPECT().EditMyInfo(gomock.Any(), gomock.Any()).Return(editedInfo, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     1,
		Data:     map[string]any{"original_description": "old desc"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *botInfoSuite) TestStep1_NoCallback_SendsHint() {
	upd := botInfoPersonalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     1,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
}

func (su *botInfoSuite) TestStep2_RestoreCallback() {
	upd := botInfoCallbackUpdate("bi:restore")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.bots.EXPECT().EditMyInfo(gomock.Any(), gomock.Any()).Return(model.BotInfo{}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{"original_description": "original"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *botInfoSuite) TestStep2_CancelCallback() {
	upd := botInfoCallbackUpdate(scenario.PayloadCancel)
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *botInfoSuite) TestStep1_EditAPIError() {
	upd := botInfoCallbackUpdate("bi:edit")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.bots.EXPECT().EditMyInfo(gomock.Any(), gomock.Any()).
		Return(model.BotInfo{}, errors.New("api error"))

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     1,
		Data:     map[string]any{"original_description": "old desc"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "edit my info")
}

func (su *botInfoSuite) TestStep1_EditSendError() {
	upd := botInfoCallbackUpdate("bi:edit")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	editedInfo := model.BotInfo{
		FirstName:   "TestBot",
		Description: "Описание изменено демо-ботом",
	}
	su.bots.EXPECT().EditMyInfo(gomock.Any(), gomock.Any()).Return(editedInfo, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     1,
		Data:     map[string]any{"original_description": "old desc"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send updated bot info")
}

func (su *botInfoSuite) TestStep1_UnknownCallback() {
	upd := botInfoCallbackUpdate("bi:unknown")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     1,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 1, state.Step)
}

func (su *botInfoSuite) TestStep1_NoCallback_HintError() {
	upd := botInfoPersonalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     1,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send hint")
}

func (su *botInfoSuite) TestStep2_NoCallback_SendsHint() {
	upd := botInfoPersonalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()
	su.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *botInfoSuite) TestStep2_NoCallback_HintError() {
	upd := botInfoPersonalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send hint")
}

func (su *botInfoSuite) TestStep2_UnknownCallback() {
	upd := botInfoCallbackUpdate("bi:unknown")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *botInfoSuite) TestStep2_RestoreAPIMissingData() {
	upd := botInfoCallbackUpdate("bi:restore")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "original_description not found")
}

func (su *botInfoSuite) TestStep2_RestoreAPIError() {
	upd := botInfoCallbackUpdate("bi:restore")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.bots.EXPECT().EditMyInfo(gomock.Any(), gomock.Any()).
		Return(model.BotInfo{}, errors.New("api error"))

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{"original_description": "original"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "restore bot info")
}

func (su *botInfoSuite) TestStep2_RestoreSendError() {
	upd := botInfoCallbackUpdate("bi:restore")
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.bots.EXPECT().EditMyInfo(gomock.Any(), gomock.Any()).Return(model.BotInfo{}, nil)
	su.ctx.EXPECT().Send(gomock.Any()).Return(errors.New("send error"))

	st := &scenario.State{
		Scenario: "botinfo",
		Step:     2,
		Data:     map[string]any{"original_description": "original"},
	}
	state, err := su.s.Handle(context.Background(), su.params(), st)

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send completion message")
}

func (su *botInfoSuite) TestUnknownStep() {
	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "botinfo",
		Step:     99,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "unknown step")
}

func TestUnitBotInfoScenario_Metadata(t *testing.T) {
	t.Parallel()

	s := scenario.BotInfoScenario{}

	assert.Equal(t, "botinfo", s.Name())
	assert.Equal(t, scenario.ChatTypePersonal, s.ChatType())
	assert.Nil(t, s.RequiredPermissions())
	assert.NotEmpty(t, s.Description())
}

// botInfoPersonalUpdate returns a personal update for botinfo tests.
func botInfoPersonalUpdate() model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 456,
		Message: &model.MessageUpdate{
			Recipient: model.Recipient{ChatType: model.ChatTypeDialog},
			Body:      model.MessageBody{Text: ""},
		},
	}
}

// botInfoCallbackUpdate returns an update with a callback payload.
func botInfoCallbackUpdate(payload string) model.Update {
	return model.Update{
		UserID:   123,
		ChatID:   456,
		Callback: &model.Callback{Payload: payload},
	}
}

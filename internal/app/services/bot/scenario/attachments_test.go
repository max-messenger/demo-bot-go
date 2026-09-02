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

func TestUnitAttachmentsScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(attachmentsSuite))
}

type attachmentsSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	api   *scenariomocks.MockAPI
	ctx   *mocks.MockContext
	msgs  *mocks.MockMessagesAPI
	chats *mocks.MockChatsAPI
	log   *zap.Logger
	s     *scenario.AttachmentsScenario
}

func (su *attachmentsSuite) SetupTest() {
	su.ctrl = gomock.NewController(su.T())
	su.api = scenariomocks.NewMockAPI(su.ctrl)
	su.ctx = mocks.NewMockContext(su.ctrl)
	su.msgs = mocks.NewMockMessagesAPI(su.ctrl)
	su.chats = mocks.NewMockChatsAPI(su.ctrl)
	su.log = zap.NewNop()
	su.s = scenario.NewAttachmentsScenario()

	apiClient := &maxbotcli.Api{
		Messages: su.msgs,
		Chats:    su.chats,
	}
	su.api.EXPECT().Client().Return(apiClient).AnyTimes()
}

func (su *attachmentsSuite) TearDownTest() {
	su.ctrl.Finish()
}

// params builds scenario.Params for the attachments scenario.
func (su *attachmentsSuite) params() scenario.Params {
	su.T().Helper()

	return scenario.Params{
		API: su.api,
		Ctx: su.ctx,
		Log: su.log,
	}
}

// personalUpdate returns a personal-chat update for attachment tests.
func personalUpdate() model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 456,
		Message: &model.MessageUpdate{
			Recipient: model.Recipient{ChatType: model.ChatTypeDialog},
			Body:      model.MessageBody{Text: ""},
		},
	}
}

func (su *attachmentsSuite) TestStep0_ImageAttachment() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(
		gomock.Any(), gomock.Any(),
	).Return(model.SendMessageResult{}, nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     0,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), "attachments", state.Scenario)
	assert.Equal(su.T(), 1, state.Step)
}

func (su *attachmentsSuite) TestStep0_SendError() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(
		gomock.Any(), gomock.Any(),
	).Return(model.SendMessageResult{}, errors.New("send failed"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     0,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send image attachment")
}

func (su *attachmentsSuite) TestStep5_Complete() {
	su.ctx.EXPECT().Send("Демо вложений завершено!").Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     5,
	})

	require.Error(su.T(), err)
	assert.ErrorIs(su.T(), err, scenario.ErrDone)
	assert.Nil(su.T(), state)
}

func (su *attachmentsSuite) TestStep5_CompleteSendError() {
	su.ctx.EXPECT().Send("Демо вложений завершено!").Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     5,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send completion message")
}

func (su *attachmentsSuite) TestUnknownStep() {
	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     99,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "unknown step")
}

func (su *attachmentsSuite) TestStep1_StickerSuccess() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	// First call: sticker message send succeeds.
	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	// Second call: explanation message with nav keyboard.
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     1,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), "attachments", state.Scenario)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *attachmentsSuite) TestStep1_StickerSendFails() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	// Sticker send fails — falls back to text-only explanation.
	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("sticker not allowed"))
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     1,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), 2, state.Step)
}

func (su *attachmentsSuite) TestStep1_StickerSendAndExplanationFail() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("sticker not allowed"))
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     1,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send sticker error explanation")
}

func (su *attachmentsSuite) TestStep1_StickerSuccessExplanationFails() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     1,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send sticker explanation")
}

func (su *attachmentsSuite) TestStep2_ContactSuccess() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	// First call: contact message send.
	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	// Second call: explanation message with nav keyboard.
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     2,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), "attachments", state.Scenario)
	assert.Equal(su.T(), 3, state.Step)
}

func (su *attachmentsSuite) TestStep2_ContactSendError() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("contact not allowed"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     2,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send contact attachment")
}

func (su *attachmentsSuite) TestStep2_ContactExplanationError() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)
	su.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(errors.New("send error"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     2,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send contact explanation")
}

func (su *attachmentsSuite) TestStep3_LocationSuccess() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     3,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), "attachments", state.Scenario)
	assert.Equal(su.T(), 4, state.Step)
}

func (su *attachmentsSuite) TestStep3_LocationSendError() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("send failed"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     3,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send location attachment")
}

func (su *attachmentsSuite) TestStep4_ShareSuccess() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, nil)

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     4,
	})

	require.NoError(su.T(), err)
	require.NotNil(su.T(), state)
	assert.Equal(su.T(), "attachments", state.Scenario)
	assert.Equal(su.T(), 5, state.Step)
}

func (su *attachmentsSuite) TestStep4_ShareSendError() {
	upd := personalUpdate()
	su.ctx.EXPECT().Update().Return(upd).AnyTimes()

	su.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).
		Return(model.SendMessageResult{}, errors.New("send failed"))

	state, err := su.s.Handle(context.Background(), su.params(), &scenario.State{
		Scenario: "attachments",
		Step:     4,
	})

	require.Error(su.T(), err)
	assert.Nil(su.T(), state)
	assert.Contains(su.T(), err.Error(), "send share attachment")
}

func TestUnitAttachmentsScenario_Metadata(t *testing.T) {
	t.Parallel()

	s := scenario.NewAttachmentsScenario()

	assert.Equal(t, "attachments", s.Name())
	assert.Equal(t, scenario.ChatTypePersonal, s.ChatType())
	assert.Nil(t, s.RequiredPermissions())
	assert.NotEmpty(t, s.Description())
}

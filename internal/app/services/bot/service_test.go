package bot_test

import (
	"context"
	"errors"
	"testing"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"demo_bot/internal/app/services/bot"
	"demo_bot/internal/app/services/bot/mocks"
	scenario "demo_bot/internal/app/services/bot/scenario"
	scenariomocks "demo_bot/internal/app/services/bot/scenario/mocks"
)

func TestUnitBotService(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(testBotService))
}

type testBotService struct {
	suite.Suite

	ctrl   *gomock.Controller
	svc    *bot.MaxBot
	client *mocks.MockClient
	store  *mocks.MockStore
	locker *mocks.MockLocker
	ctx    *mocks.MockContext
	api    *scenariomocks.MockAPI
	msgs   *mocks.MockMessagesAPI
	chats  *mocks.MockChatsAPI
}

func (s *testBotService) SetupTest() {
	s.ctrl = gomock.NewController(s.T())

	s.client = mocks.NewMockClient(s.ctrl)
	s.store = mocks.NewMockStore(s.ctrl)
	s.locker = mocks.NewMockLocker(s.ctrl)
	s.ctx = mocks.NewMockContext(s.ctrl)

	s.api = scenariomocks.NewMockAPI(s.ctrl)
	s.msgs = mocks.NewMockMessagesAPI(s.ctrl)
	s.chats = mocks.NewMockChatsAPI(s.ctrl)

	apiClient := &maxbotcli.Api{
		Messages: s.msgs,
		Chats:    s.chats,
	}
	s.api.EXPECT().Client().Return(apiClient).AnyTimes()

	s.svc = bot.New(
		zap.NewNop(),
		s.client,
		s.api,
		s.store,
		s.locker,
	)
}

func (s *testBotService) TearDownTest() {
	s.ctrl.Finish()
}

// capturedHandler stores a handler with its optional middleware.
type capturedHandler struct {
	handler    maxbot.HandlerFunc
	middleware maxbot.MiddlewareFunc
}

// invoke calls the handler through its middleware (if present).
func (c capturedHandler) invoke(ctx maxbot.Context) error {
	if c.middleware != nil {
		return c.middleware(c.handler)(ctx)
	}

	return c.handler(ctx)
}

// startAndCapture calls Start() and captures all registered handlers.
func (s *testBotService) startAndCapture() map[string]capturedHandler {
	s.T().Helper()

	handlers := make(map[string]capturedHandler)

	s.client.EXPECT().Handle(maxbot.OnBotStarted, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ep string, h maxbot.HandlerFunc, m ...maxbot.MiddlewareFunc) {
			var mw maxbot.MiddlewareFunc
			if len(m) > 0 {
				mw = m[0]
			}

			handlers[ep] = capturedHandler{handler: h, middleware: mw}
		})
	s.client.EXPECT().Handle(maxbot.OnMessageCreated, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ep string, h maxbot.HandlerFunc, m ...maxbot.MiddlewareFunc) {
			var mw maxbot.MiddlewareFunc
			if len(m) > 0 {
				mw = m[0]
			}

			handlers[ep] = capturedHandler{handler: h, middleware: mw}
		})
	s.client.EXPECT().Handle(maxbot.OnMessageCallback, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ep string, h maxbot.HandlerFunc, m ...maxbot.MiddlewareFunc) {
			var mw maxbot.MiddlewareFunc
			if len(m) > 0 {
				mw = m[0]
			}

			handlers[ep] = capturedHandler{handler: h, middleware: mw}
		})
	s.client.EXPECT().Handle(maxbot.OnBotAdded, gomock.Any()).
		DoAndReturn(func(ep string, h maxbot.HandlerFunc, _ ...maxbot.MiddlewareFunc) {
			handlers[ep] = capturedHandler{handler: h}
		})
	s.client.EXPECT().Handle(maxbot.OnUserAdded, gomock.Any()).
		DoAndReturn(func(ep string, h maxbot.HandlerFunc, _ ...maxbot.MiddlewareFunc) {
			handlers[ep] = capturedHandler{handler: h}
		})
	s.client.EXPECT().Handle(maxbot.OnChatTitleChangedEvent, gomock.Any()).
		DoAndReturn(func(ep string, h maxbot.HandlerFunc, _ ...maxbot.MiddlewareFunc) {
			handlers[ep] = capturedHandler{handler: h}
		})

	_ = s.svc.Start(context.Background())

	return handlers
}

// expectLock sets up the lock/unlock expectations for stateMiddleware.
func (s *testBotService) expectLock() {
	s.T().Helper()

	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.locker.EXPECT().Lock(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(func(_ context.Context) error { return nil }, nil)
}

// --- Lifecycle ---

func (s *testBotService) TestStart() {
	s.startAndCapture()
}

func (s *testBotService) TestStop() {
	err := s.svc.Stop(context.Background())
	s.NoError(err)
}

// --- startHandler ---

func (s *testBotService) TestStartHandler() {
	handlers := s.startAndCapture()

	s.ctx.EXPECT().Update().Return(personalUpdate()).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.expectLock()

	err := handlers[maxbot.OnBotStarted].invoke(s.ctx)
	s.NoError(err)
}

// --- messageHandler ---

func (s *testBotService) TestMessageHandler_NoState_DispatchStart() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "/start"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestMessageHandler_NoState_DispatchHelp() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "/help"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestMessageHandler_NoState_UnknownCommand() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "/unknown"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestMessageHandler_NoState_EmptyCommand() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "hello"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestMessageHandler_HasState() {
	handlers := s.startAndCapture()

	// Leave scenario at step 0 with text message shows confirmation,
	// then returns a new state → SetState is called.
	upd := groupUpdate()
	upd.Message.Body.Text = "some text"
	state := &scenario.State{Scenario: "leave", Step: 0}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(state, nil)
	s.store.EXPECT().SetState(gomock.Any(), upd.UserID, gomock.Any()).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestMessageHandler_StoreError() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "test"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, errors.New("db error"))

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.Error(err)
}

// --- callbackHandler ---

func (s *testBotService) TestCallbackHandler_CmdStart() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: scenario.PayloadCmdStart}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestCallbackHandler_CmdHelp() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: scenario.PayloadCmdHelp}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(nil)
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestCallbackHandler_StartScenario() {
	handlers := s.startAndCapture()

	// Starting a scenario: permission check + Handle step 0 + SetState.
	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: "s:message"}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	// Message scenario step 0 uses Messages.Send and Ctx.Send.
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.store.EXPECT().SetState(gomock.Any(), upd.UserID, gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestCallbackHandler_Cancel() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: scenario.PayloadCancel}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestCallbackHandler_StartScenarioNotFound() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: "s:nonexistent"}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.ctx.EXPECT().Send(gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestCallbackHandler_WithState() {
	handlers := s.startAndCapture()

	// Leave scenario step 0, callback "Нет" → sends "Отменено", then ErrDone
	// → ClearState + sendStartMenu (another Send with keyboard).
	upd := groupUpdate()
	upd.Callback = &model.Callback{Payload: "leave:no"}
	upd.Message = nil
	state := &scenario.State{Scenario: "leave", Step: 0}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(state, nil)
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any()).Return(nil)             // "Отменено"
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil) // start menu

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

// --- dispatchCommand: wrong chat type ---

func (s *testBotService) TestDispatchCommand_Scenario_WrongChatType() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "/chat"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

// --- handleScenarioStep ---

func (s *testBotService) TestHandleScenarioStep_ErrDone() {
	handlers := s.startAndCapture()

	// Leave scenario with "Нет" callback: Send("Отменено"), ErrDone →
	// ClearState + sendStartMenu.
	upd := groupUpdate()
	upd.Callback = &model.Callback{Payload: "leave:no"}
	upd.Message = nil
	state := &scenario.State{Scenario: "leave", Step: 0}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(state, nil)
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any()).Return(nil)             // "Отменено"
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil) // start menu

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestHandleScenarioStep_UnknownScenario() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "text"
	state := &scenario.State{Scenario: "nonexistent", Step: 1}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(state, nil)
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

// --- stateMiddleware ---

func (s *testBotService) TestStateMiddleware_LockError() {
	handlers := s.startAndCapture()

	upd := personalUpdate()

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.locker.EXPECT().Lock(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(nil, errors.New("lock error"))

	err := handlers[maxbot.OnBotStarted].invoke(s.ctx)
	s.Error(err)
}

// --- Group handlers ---

func (s *testBotService) TestBotAddedHandler() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnBotAdded].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestUserAddedHandler() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.User = &model.User{FirstName: "Test"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnUserAdded].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestChatTitleChangedHandler_EmptyTitle() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.ChatProp = &model.ChatProp{Title: ""}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()

	err := handlers[maxbot.OnChatTitleChangedEvent].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestChatTitleChangedHandler_WithTitle() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.ChatProp = &model.ChatProp{Title: "New Title"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnChatTitleChangedEvent].invoke(s.ctx)
	s.NoError(err)
}

// --- checkPermissions ---

func (s *testBotService) TestCheckPermissions_MissingPerms() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.Callback = &model.Callback{Payload: "s:admins"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.chats.EXPECT().GetMembership(gomock.Any(), upd.ChatID).Return(model.ChatMember{
		Permissions: []model.ChatAdminPermission{},
	}, nil)
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestCheckPermissions_AllPresent() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.Callback = &model.Callback{Payload: "s:admins"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.chats.EXPECT().GetMembership(gomock.Any(), upd.ChatID).Return(model.ChatMember{
		Permissions: []model.ChatAdminPermission{"add_admins"},
	}, nil)
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)
	s.chats.EXPECT().GetAdmins(gomock.Any(), upd.ChatID).Return(model.ChatMembersList{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.store.EXPECT().SetState(gomock.Any(), upd.UserID, gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

// --- handleScenarioStep: ErrDoneNoMenu ---

func (s *testBotService) TestHandleScenarioStep_ErrDoneNoMenu() {
	handlers := s.startAndCapture()

	// Leave scenario "yes" callback: LeaveChat + ErrDoneNoMenu → ClearState, no menu.
	upd := groupUpdate()
	upd.Callback = &model.Callback{Payload: "leave:yes"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(&scenario.State{Scenario: "leave", Step: 0}, nil)
	s.chats.EXPECT().LeaveChat(gomock.Any(), upd.ChatID).Return(model.SimpleQueryResult{Success: true}, nil)
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

// --- handleScenarioStep: ClearState error ---

func (s *testBotService) TestHandleScenarioStep_ClearStateError() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "text"
	state := &scenario.State{Scenario: "nonexistent", Step: 1}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(state, nil)
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(errors.New("db error"))

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "clear state")
}

// --- dispatchCommand: correct chat type ---

func (s *testBotService) TestDispatchCommand_Scenario_CorrectChatType() {
	handlers := s.startAndCapture()

	// /message from personal chat — ChatTypePersonal scenario, allowed.
	upd := personalUpdate()
	upd.Message.Body.Text = "/message"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	s.store.EXPECT().SetState(gomock.Any(), upd.UserID, gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

// --- dispatchCommand: /help from personal chat ---

func (s *testBotService) TestDispatchCommand_HelpPersonal() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Message.Body.Text = "/help"

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnMessageCreated].invoke(s.ctx)
	s.NoError(err)
}

// --- sendStartMenu: group chat ---

func (s *testBotService) TestSendStartMenu_Group() {
	handlers := s.startAndCapture()

	upd := groupUpdate()

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnBotStarted].invoke(s.ctx)
	s.NoError(err)
}

// --- callbackHandler: cancel ClearState error ---

func (s *testBotService) TestCallbackHandler_CancelClearStateError() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: scenario.PayloadCancel}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(errors.New("db error"))

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "clear state")
}

// --- callbackHandler: non-command, no state -> sendStartMenu ---

func (s *testBotService) TestCallbackHandler_NonCommandNoState() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: "some:other"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().GetState(gomock.Any(), upd.UserID).Return(nil, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

// --- checkPermissions: partial permissions ---

func (s *testBotService) TestCheckPermissions_PartialPerms() {
	handlers := s.startAndCapture()

	// chat scenario requires "pin_message" and "change_chat_info".
	// Bot has only "pin_message" — missing "change_chat_info".
	upd := groupUpdate()
	upd.Callback = &model.Callback{Payload: "s:chat"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.chats.EXPECT().GetMembership(gomock.Any(), upd.ChatID).Return(model.ChatMember{
		Permissions: []model.ChatAdminPermission{"pin_message"},
	}, nil)
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.NoError(err)
}

// --- checkPermissions: GetMembership API error ---

func (s *testBotService) TestCheckPermissions_GetMembershipError() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.Callback = &model.Callback{Payload: "s:admins"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.chats.EXPECT().GetMembership(gomock.Any(), upd.ChatID).Return(model.ChatMember{}, errors.New("api error"))

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "check permissions")
}

// --- Group handlers: error paths ---

func (s *testBotService) TestBotAddedHandler_SendError() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, errors.New("send error"))

	err := handlers[maxbot.OnBotAdded].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "send bot_added welcome")
}

func (s *testBotService) TestUserAddedHandler_FallbackName() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.User = nil
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnUserAdded].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestUserAddedHandler_EmptyName() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.User = &model.User{FirstName: ""}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)

	err := handlers[maxbot.OnUserAdded].invoke(s.ctx)
	s.NoError(err)
}

func (s *testBotService) TestUserAddedHandler_SendError() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.User = &model.User{FirstName: "Test"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, errors.New("send error"))

	err := handlers[maxbot.OnUserAdded].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "send user_added welcome")
}

func (s *testBotService) TestChatTitleChangedHandler_SendError() {
	handlers := s.startAndCapture()

	upd := groupUpdate()
	upd.ChatProp = &model.ChatProp{Title: "New Title"}
	upd.Message = nil

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.msgs.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, errors.New("send error"))

	err := handlers[maxbot.OnChatTitleChangedEvent].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "send title changed notification")
}

// --- handleCommandCallback: ClearState error paths ---

func (s *testBotService) TestHandleCommandCallback_CmdStart_ClearStateError() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: scenario.PayloadCmdStart}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(errors.New("db error"))

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "clear state on cmd:start")
}

func (s *testBotService) TestHandleCommandCallback_CmdHelp_ClearStateError() {
	handlers := s.startAndCapture()

	upd := personalUpdate()
	upd.Callback = &model.Callback{Payload: scenario.PayloadCmdHelp}

	s.ctx.EXPECT().Update().Return(upd).AnyTimes()
	s.ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	s.expectLock()
	s.store.EXPECT().ClearState(gomock.Any(), upd.UserID).Return(errors.New("db error"))

	err := handlers[maxbot.OnMessageCallback].invoke(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "clear state on cmd:help")
}

// --- Update helpers ---

func personalUpdate() model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 456,
		Message: &model.MessageUpdate{
			Recipient: model.Recipient{
				ChatType: model.ChatTypeDialog,
			},
			Body: model.MessageBody{Text: ""},
		},
	}
}

func groupUpdate() model.Update {
	return model.Update{
		UserID: 123,
		ChatID: 789,
		Message: &model.MessageUpdate{
			Recipient: model.Recipient{
				ChatType: model.ChatTypeChat,
			},
			Body: model.MessageBody{Text: ""},
		},
	}
}

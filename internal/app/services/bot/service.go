package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/max-messenger/maxbot"
	"go.uber.org/zap"

	"demo_bot/internal/app/services/bot/scenario"
	"demo_bot/pkg/locker"
)

const moduleName = "demo_bot"

//go:generate go tool mockgen -source=service.go -destination=mocks/service_mocks.go -package=mocks

// Store defines the state persistence interface.
type Store interface {
	GetState(ctx context.Context, userID int64) (*scenario.State, error)
	SetState(ctx context.Context, userID int64, state *scenario.State) error
	ClearState(ctx context.Context, userID int64) error
}

// Client defines the bot framework routing interface.
type Client interface {
	Handle(endpoint string, h maxbot.HandlerFunc, m ...maxbot.MiddlewareFunc)
	HandleCallback(endpoint string, h maxbot.HandlerFunc, m ...maxbot.MiddlewareFunc)
}

// API defines the bot API client interface.
type API = scenario.API

// Locker defines the distributed lock interface.
type Locker interface {
	Lock(ctx context.Context, key locker.Key, value string, ttl time.Duration) (func(ctx context.Context) error, error)
}

// MaxBot is the core bot service that dispatches updates to scenarios.
type MaxBot struct {
	log      *zap.Logger
	bot      Client
	api      API
	store    Store
	locker   Locker
	registry *scenario.Registry
}

// New creates a new MaxBot service with all registered scenarios.
func New(l *zap.Logger, c Client, a API, s Store, lr Locker) *MaxBot {
	reg := scenario.NewRegistry()
	reg.Add(scenario.MessageScenario{})
	reg.Add(scenario.ActionsScenario{})
	reg.Add(scenario.KeyboardScenario{})
	reg.Add(scenario.AttachmentsScenario{})
	reg.Add(scenario.UploadScenario{})
	reg.Add(scenario.BotInfoScenario{})
	reg.Add(scenario.ChatScenario{})
	reg.Add(scenario.MembersScenario{})
	reg.Add(scenario.AdminsScenario{})
	reg.Add(scenario.LeaveScenario{})
	reg.Add(scenario.MentionScenario{})

	return &MaxBot{
		bot:      c,
		api:      a,
		log:      l,
		store:    s,
		locker:   lr,
		registry: reg,
	}
}

// Start registers all bot handlers.
func (m *MaxBot) Start(_ context.Context) error {
	m.bot.Handle(maxbot.OnBotStarted, m.startHandler, m.stateMiddleware)
	m.bot.Handle(maxbot.OnMessageCreated, m.messageHandler, m.stateMiddleware)
	m.bot.Handle(maxbot.OnMessageCallback, m.callbackHandler, m.stateMiddleware)

	// Group event handlers (no stateMiddleware — no user state flow).
	m.bot.Handle(maxbot.OnBotAdded, m.botAddedHandler)
	m.bot.Handle(maxbot.OnUserAdded, m.userAddedHandler)
	m.bot.Handle(maxbot.OnChatTitleChangedEvent, m.chatTitleChangedHandler)

	return nil
}

// Stop is a no-op for the bot service.
func (m *MaxBot) Stop(_ context.Context) error {
	return nil
}

func (m *MaxBot) startHandler(ctx maxbot.Context) error {
	return m.sendStartMenu(ctx)
}

func (m *MaxBot) messageHandler(ctx maxbot.Context) error {
	upd := ctx.Update()
	userID := upd.UserID

	state, err := m.store.GetState(ctx.Context(), userID)
	if err != nil {
		return fmt.Errorf("get state: %w", err)
	}

	if state != nil {
		return m.handleScenarioStep(ctx, state)
	}

	return m.dispatchCommand(ctx)
}

func (m *MaxBot) callbackHandler(ctx maxbot.Context) error {
	upd := ctx.Update()
	payload := upd.Callback.Payload
	userID := upd.UserID

	// Handle command buttons (/start, /help).
	if handled, err := m.handleCommandCallback(ctx, payload, userID); handled {
		return err
	}

	if strings.HasPrefix(payload, "s:") {
		scenarioName := strings.TrimPrefix(payload, "s:")
		s := m.registry.Get(scenarioName)
		if s == nil {
			return ctx.Send(fmt.Sprintf("Сценарий /%s не найден", scenarioName))
		}

		return m.startScenario(ctx, s)
	}

	if payload == scenario.PayloadCancel {
		err := m.store.ClearState(ctx.Context(), userID)
		if err != nil {
			return fmt.Errorf("clear state: %w", err)
		}

		return m.sendStartMenu(ctx)
	}

	state, err := m.store.GetState(ctx.Context(), userID)
	if err != nil {
		return fmt.Errorf("get state: %w", err)
	}

	if state != nil {
		return m.handleScenarioStep(ctx, state)
	}

	return m.sendStartMenu(ctx)
}

func (m *MaxBot) dispatchCommand(ctx maxbot.Context) error {
	upd := ctx.Update()
	cmd := maxbotcli.GetCommand(upd)

	switch cmd {
	case "/start":
		return m.sendStartMenu(ctx)
	case "/help":
		return m.sendHelpText(ctx)
	default:
		if cmd != "" {
			s := m.registry.Get(strings.TrimPrefix(cmd, "/"))
			if s != nil {
				if !m.isChatTypeAllowed(s, ctx) {
					return ctx.Send(fmt.Sprintf("Сценарий /%s доступен только в %s чате", s.Name(), s.ChatType()))
				}

				return m.startScenario(ctx, s)
			}

			return ctx.Send(fmt.Sprintf("Неизвестная команда: %s\nОтправьте /help для справки", cmd),
				maxbot.WithKeyboard(scenario.MenuKeyboard()))
		}

		// Plain text with no active scenario — ignore silently.
		return nil
	}
}

func (m *MaxBot) startScenario(ctx maxbot.Context, s scenario.Scenario) error {
	// Check permissions for group scenarios before starting.
	if s.ChatType() == scenario.ChatTypeGroup && len(s.RequiredPermissions()) > 0 {
		missing, err := m.checkPermissions(ctx, s)
		if err != nil {
			return fmt.Errorf("check permissions: %w", err)
		}

		if len(missing) > 0 {
			return m.sendPermissionWarning(ctx, s, missing)
		}

		// Permissions are present — show brief notice.
		if err := m.sendPermissionNotice(ctx, s); err != nil {
			return fmt.Errorf("send permission notice: %w", err)
		}
	}

	p := scenario.Params{
		API: m.api,
		Ctx: ctx,
		Log: m.log,
	}

	state := &scenario.State{
		Scenario: s.Name(),
		Step:     0,
	}

	newState, err := s.Handle(ctx.Context(), p, state)
	if err != nil {
		return fmt.Errorf("scenario %s handle step 0: %w", s.Name(), err)
	}

	if newState == nil {
		return m.sendStartMenu(ctx)
	}

	return m.store.SetState(ctx.Context(), ctx.Update().UserID, newState)
}

func (m *MaxBot) handleScenarioStep(ctx maxbot.Context, state *scenario.State) error {
	s := m.registry.Get(state.Scenario)
	if s == nil {
		err := m.store.ClearState(ctx.Context(), ctx.Update().UserID)
		if err != nil {
			return fmt.Errorf("clear state: %w", err)
		}

		return m.sendStartMenu(ctx)
	}

	p := scenario.Params{
		API: m.api,
		Ctx: ctx,
		Log: m.log,
	}

	newState, err := s.Handle(ctx.Context(), p, state)
	if err != nil {
		if errors.Is(err, scenario.ErrDone) || errors.Is(err, scenario.ErrDoneNoMenu) {
			if clearErr := m.store.ClearState(ctx.Context(), ctx.Update().UserID); clearErr != nil {
				return fmt.Errorf("clear state: %w", clearErr)
			}

			if errors.Is(err, scenario.ErrDoneNoMenu) {
				return nil
			}

			return m.sendStartMenu(ctx)
		}

		return fmt.Errorf("scenario %s handle step %d: %w", s.Name(), state.Step, err)
	}

	if newState == nil {
		err = m.store.ClearState(ctx.Context(), ctx.Update().UserID)
		if err != nil {
			return fmt.Errorf("clear state: %w", err)
		}

		return m.sendStartMenu(ctx)
	}

	return m.store.SetState(ctx.Context(), ctx.Update().UserID, newState)
}

func (m *MaxBot) sendStartMenu(ctx maxbot.Context) error {
	chatType, p := m.chatTypeAndParams(ctx)
	isPersonal := chatType == scenario.ChatTypePersonal
	scenarios := m.registry.ListByChatType(chatType)

	return scenario.SendStartMenu(p, scenarios, isPersonal)
}

func (m *MaxBot) sendHelpText(ctx maxbot.Context) error {
	chatType, p := m.chatTypeAndParams(ctx)
	isPersonal := chatType == scenario.ChatTypePersonal
	scenarios := m.registry.ListByChatType(chatType)

	if isPersonal {
		return scenario.SendHelpText(p, scenarios)
	}

	return scenario.SendGroupHelpText(p, scenarios)
}

func (m *MaxBot) chatTypeAndParams(ctx maxbot.Context) (scenario.ChatType, scenario.Params) {
	upd := ctx.Update()
	isPersonal := scenario.IsPersonalChat(upd)

	chatType := scenario.ChatTypePersonal
	if !isPersonal {
		chatType = scenario.ChatTypeGroup
	}

	p := scenario.Params{Ctx: ctx, Log: m.log, API: m.api}

	return chatType, p
}

func (m *MaxBot) isChatTypeAllowed(s scenario.Scenario, ctx maxbot.Context) bool {
	upd := ctx.Update()

	if s.ChatType() == scenario.ChatTypeAny {
		return true
	}

	if s.ChatType() == scenario.ChatTypePersonal && scenario.IsPersonalChat(upd) {
		return true
	}

	if s.ChatType() == scenario.ChatTypeGroup && scenario.IsGroupChat(upd) {
		return true
	}

	return false
}

func (m *MaxBot) stateMiddleware(next maxbot.HandlerFunc) maxbot.HandlerFunc {
	return func(ctx maxbot.Context) error {
		unlock, err := m.lock(ctx.Context(), ctx.Update().UserID)
		if err != nil {
			return err
		}

		defer func() {
			if unlockErr := unlock(ctx.Context()); unlockErr != nil {
				m.log.Error("cannot unlock user", zap.Error(unlockErr))
			}
		}()

		return next(ctx)
	}
}

func (m *MaxBot) lock(ctx context.Context, userID int64) (func(ctx context.Context) error, error) {
	lockKey := locker.Key{
		Component: moduleName,
		Key:       fmt.Sprintf("user-lock:%d", userID),
	}

	return m.locker.Lock(ctx, lockKey, uuid.NewString(), time.Second*10)
}

func (m *MaxBot) checkPermissions(ctx maxbot.Context, s scenario.Scenario) (missing []string, err error) {
	permissions := s.RequiredPermissions()
	if len(permissions) == 0 {
		return nil, nil
	}

	membership, err := m.api.Client().Chats.GetMembership(ctx.Context(), ctx.Update().ChatID)
	if err != nil {
		return nil, fmt.Errorf("get membership: %w", err)
	}

	botPerms := make(map[string]bool, len(membership.Permissions))
	for _, p := range membership.Permissions {
		botPerms[string(p)] = true
	}

	for _, req := range permissions {
		if !botPerms[req] {
			missing = append(missing, req)
		}
	}

	return missing, nil
}

func (m *MaxBot) sendPermissionWarning(ctx maxbot.Context, s scenario.Scenario, missing []string) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Сценарий /%s требует следующих прав, которых у бота нет:\n\n", s.Name())

	for _, perm := range missing {
		fmt.Fprintf(&sb, "- **%s**\n", perm)
	}

	sb.WriteString("\nНазначьте бота администратором с нужными правами и повторите.")

	upd := ctx.Update()

	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		SetText(sb.String()).
		SetFormat(model.FormatMarkdown).
		AddKeyboard(scenario.MenuKeyboard())

	_, err := m.api.Client().Messages.Send(ctx.Context(), msg)
	if err != nil {
		return fmt.Errorf("send permission warning: %w", err)
	}

	return nil
}

func (m *MaxBot) sendPermissionNotice(ctx maxbot.Context, s scenario.Scenario) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Для сценария /%s нужны права: ", s.Name())

	for i, perm := range s.RequiredPermissions() {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "**%s**", perm)
	}

	sb.WriteString(" — у бота они есть.")

	upd := ctx.Update()

	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		SetText(sb.String()).
		SetFormat(model.FormatMarkdown)

	_, err := m.api.Client().Messages.Send(ctx.Context(), msg)
	if err != nil {
		return fmt.Errorf("send permission notice: %w", err)
	}

	return nil
}

func (m *MaxBot) botAddedHandler(ctx maxbot.Context) error {
	upd := ctx.Update()

	msg := maxbotcli.NewMessage().
		SetChat(upd.ChatID).
		SetText("Привет! Я бот-демонстратор API MAX.\n\n" +
			"Отправьте /help для списка команд или /start для меню.").
		WithoutNotify().
		AddKeyboard(scenario.MenuKeyboard())

	if _, err := m.api.Client().Messages.Send(ctx.Context(), msg); err != nil {
		return fmt.Errorf("send bot_added welcome: %w", err)
	}

	return nil
}

func (m *MaxBot) userAddedHandler(ctx maxbot.Context) error {
	upd := ctx.Update()

	userName := "новый участник"
	if upd.User != nil && upd.User.Name != "" {
		userName = upd.User.Name
	}

	msg := maxbotcli.NewMessage().
		SetChat(upd.ChatID).
		SetText(fmt.Sprintf("Добро пожаловать, %s! Отправьте /help для списка команд.", userName)).
		WithoutNotify().
		AddKeyboard(scenario.MenuKeyboard())

	if _, err := m.api.Client().Messages.Send(ctx.Context(), msg); err != nil {
		return fmt.Errorf("send user_added welcome: %w", err)
	}

	return nil
}

func (m *MaxBot) chatTitleChangedHandler(ctx maxbot.Context) error {
	upd := ctx.Update()
	chatProp := upd.GetChat()

	if chatProp.Title == "" {
		return nil
	}

	msg := maxbotcli.NewMessage().
		SetChat(upd.ChatID).
		SetText(fmt.Sprintf("Название чата изменено на: %s", chatProp.Title)).
		WithoutNotify()

	if _, err := m.api.Client().Messages.Send(ctx.Context(), msg); err != nil {
		return fmt.Errorf("send title changed notification: %w", err)
	}

	return nil
}

// handleCommandCallback handles /start and /help callback buttons.
// Returns (true, err) if the payload was a command callback, (false, nil) otherwise.
func (m *MaxBot) handleCommandCallback(ctx maxbot.Context, payload string, userID int64) (bool, error) {
	switch payload {
	case scenario.PayloadCmdStart:
		if err := m.store.ClearState(ctx.Context(), userID); err != nil {
			return true, fmt.Errorf("clear state on cmd:start: %w", err)
		}

		return true, m.sendStartMenu(ctx)
	case scenario.PayloadCmdHelp:
		if err := m.store.ClearState(ctx.Context(), userID); err != nil {
			return true, fmt.Errorf("clear state on cmd:help: %w", err)
		}

		return true, m.sendHelpText(ctx)
	default:
		return false, nil
	}
}

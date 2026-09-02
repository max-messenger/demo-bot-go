package scenario

import (
	"context"
	"fmt"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
)

const (
	keyboardName      = "keyboard"
	payloadKBCallback = "kb:callback"
)

// KeyboardScenario demonstrates all keyboard and button types available in MAX Bot API.
type KeyboardScenario struct{}

// Name returns the scenario command name.
func (KeyboardScenario) Name() string { return keyboardName }

// Description returns a short Russian description for /help.
func (KeyboardScenario) Description() string { return "демо клавиатур и кнопок" }

// ChatType returns Personal — this scenario only works in direct messages.
func (KeyboardScenario) ChatType() ChatType { return ChatTypePersonal }

// RequiredPermissions returns nil — no admin permissions needed for personal chat.
func (KeyboardScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the keyboard demo scenario.
func (KeyboardScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	switch s.Step {
	case 0:
		return handleKBStep0(ctx, p, s)
	case 1:
		return handleKBStep1(p)
	case 2:
		return handleKBStep2(p)
	case 3:
		return handleKBStep3(p)
	case 4:
		return handleKBStep4(p)
	case 5:
		return handleKBStep5(p)
	case 6:
		return handleKBStep6(p)
	default:
		return nil, ErrDone
	}
}

// handleKBStep0: callback button — wait for user to click it.
func handleKBStep0(ctx context.Context, p Params, _ *State) (*State, error) {
	upd := p.Ctx.Update()

	if upd.Callback != nil && upd.Callback.Payload == payloadKBCallback {
		notification := "Вы нажали callback-кнопку!"
		_, err := p.API.Client().Messages.AnswerOnCallback(ctx, upd.Callback.CallbackID, model.CallbackAnswer{
			Notification: &notification,
		})
		if err != nil {
			return nil, fmt.Errorf("answer on callback: %w", err)
		}

		err = p.Ctx.Send(
			"AnswerOnCallback отправляет уведомление, которое видит только пользователь, нажавший кнопку.",
			maxbot.WithKeyboard(navKeyboard()),
		)
		if err != nil {
			return nil, fmt.Errorf("send explanation: %w", err)
		}

		return &State{Scenario: keyboardName, Step: 1}, nil
	}

	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("Нажми меня", payloadKBCallback)
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	err := p.Ctx.Send(
		"1. Callback-кнопка — при нажатии бот получает событие callback и может ответить уведомлением. "+
			"Нажмите кнопку ниже:",
		maxbot.WithKeyboard(kb),
	)
	if err != nil {
		return nil, fmt.Errorf("send callback demo: %w", err)
	}

	return &State{Scenario: keyboardName, Step: 0}, nil
}

// handleKBStep1: link button.
func handleKBStep1(p Params) (*State, error) {
	kb := model.NewKeyboard()
	kb.AddRow().AddLink("Открыть dev портал MAX", "https://dev.max.ru/docs")
	kb.AddRow().AddCallBack("Далее", "next")
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	err := p.Ctx.Send(
		"2. Link-кнопка — открывает URL в браузере. Нажмите кнопку ниже:",
		maxbot.WithKeyboard(kb),
	)
	if err != nil {
		return nil, fmt.Errorf("send link demo: %w", err)
	}

	return &State{Scenario: keyboardName, Step: 2}, nil
}

// handleKBStep2: reply keyboard with geo-location request button.
func handleKBStep2(p Params) (*State, error) {
	kb := replyKeyboard()
	kb.AddRow().AddGeoLocation("Отправить геолокацию", true)

	if err := p.Ctx.Send(
		"3. Кнопка запроса геолокации (reply_keyboard) — предлагает пользователю поделиться "+
			"своим местоположением. Нажмите кнопку ниже:",
		maxbot.WithKeyboard(kb),
	); err != nil {
		return nil, fmt.Errorf("send geo keyboard: %w", err)
	}

	if err := sendWithNav(p, "Попробуйте нажать кнопку геолокации выше, или нажмите «Далее»."); err != nil {
		return nil, fmt.Errorf("send nav: %w", err)
	}

	return &State{Scenario: keyboardName, Step: 3}, nil
}

// handleKBStep3: reply keyboard with contact request button.
func handleKBStep3(p Params) (*State, error) {
	kb := replyKeyboard()
	kb.AddRow().AddContact("Отправить контакт")

	if err := p.Ctx.Send(
		"4. Кнопка запроса контакта (reply_keyboard) — предлагает пользователю поделиться "+
			"своим контактом. Нажмите кнопку ниже:",
		maxbot.WithKeyboard(kb),
	); err != nil {
		return nil, fmt.Errorf("send contact keyboard: %w", err)
	}

	if err := sendWithNav(p, "Попробуйте нажать кнопку контакта выше, или нажмите «Далее»."); err != nil {
		return nil, fmt.Errorf("send nav: %w", err)
	}

	return &State{Scenario: keyboardName, Step: 4}, nil
}

// handleKBStep4: clipboard button (inline keyboard).
func handleKBStep4(p Params) (*State, error) {
	kb := model.NewKeyboard()
	kb.AddRow().AddClipboard("Скопировать текст", "скопированный текст")
	kb.AddRow().AddCallBack("Далее", "next")
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	err := p.Ctx.Send(
		"5. Clipboard-кнопка — копирует указанный текст в буфер обмена пользователя. "+
			"Нажмите кнопку ниже:",
		maxbot.WithKeyboard(kb),
	)
	if err != nil {
		return nil, fmt.Errorf("send clipboard demo: %w", err)
	}

	return &State{Scenario: keyboardName, Step: 5}, nil
}

// handleKBStep5: reply keyboard with message button.
func handleKBStep5(p Params) (*State, error) {
	kb := replyKeyboard()
	kb.AddRow().AddMessage("Привет от кнопки!")

	if err := p.Ctx.Send(
		"6. Кнопка отправки сообщения (reply_keyboard) — при нажатии отправляет "+
			"текст от имени пользователя в чат. Нажмите кнопку ниже:",
		maxbot.WithKeyboard(kb),
	); err != nil {
		return nil, fmt.Errorf("send message button keyboard: %w", err)
	}

	if err := p.Ctx.Send("Попробуйте нажать кнопку выше — она отправит сообщение от вашего имени.",
		maxbot.WithKeyboard(navKeyboard())); err != nil {
		return nil, fmt.Errorf("send nav: %w", err)
	}

	return &State{Scenario: keyboardName, Step: 6}, nil
}

// handleKBStep6: demo complete.
func handleKBStep6(p Params) (*State, error) {
	if err := p.Ctx.Send("Демо клавиатур завершено!"); err != nil {
		return nil, fmt.Errorf("send final message: %w", err)
	}

	return nil, ErrDone
}

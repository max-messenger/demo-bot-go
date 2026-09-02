package scenario

import (
	"context"
	"fmt"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"go.uber.org/zap"
)

// AttachmentsScenario demonstrates all attachment types: image, sticker, contact, location, share.
type AttachmentsScenario struct{}

// NewAttachmentsScenario creates a new attachments scenario.
func NewAttachmentsScenario() *AttachmentsScenario {
	return &AttachmentsScenario{}
}

// Name returns the scenario command name.
func (AttachmentsScenario) Name() string { return "attachments" }

// Description returns a short Russian description for /help.
func (AttachmentsScenario) Description() string {
	return "демо вложений: изображение, стикер, контакт, локация, ссылка"
}

// ChatType returns where this scenario can run.
func (AttachmentsScenario) ChatType() ChatType { return ChatTypePersonal }

// RequiredPermissions returns admin permissions needed (group only).
func (AttachmentsScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the attachments scenario.
func (a AttachmentsScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	switch s.Step {
	case 0:
		return a.stepImage(ctx, p)
	case 1:
		return a.stepSticker(ctx, p)
	case 2:
		return a.stepContact(ctx, p)
	case 3:
		return a.stepLocation(ctx, p)
	case 4:
		return a.stepShare(ctx, p)
	case 5:
		return a.stepComplete(p)
	default:
		return nil, fmt.Errorf("attachments: unknown step %d", s.Step)
	}
}

func (a AttachmentsScenario) stepImage(ctx context.Context, p Params) (*State, error) {
	upd := p.Ctx.Update()

	const text = "1. Вложение: изображение (AddImageUrl)\n\n" +
		"Метод msg.AddImageUrl(url) прикрепляет изображение по URL к сообщению."

	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		SetText(text).
		AddImageUrl("https://img.icons8.com/clouds/200/chat.png")

	msg.AddKeyboard(navKeyboard())

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("send image attachment: %w", err)
	}

	return &State{Scenario: a.Name(), Step: 1}, nil
}

func (a AttachmentsScenario) stepSticker(ctx context.Context, p Params) (*State, error) {
	const text = "2. Вложение: стикер (AddSticker)\n\n" +
		"Метод msg.AddSticker(code) прикрепляет стикер по его коду. " +
		"Стикер должен быть единственным вложением в сообщении."

	upd := p.Ctx.Update()

	// Sticker must be the only attachment — send without text.
	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		AddSticker("1d3acfe92")

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		p.Log.Warn("sticker send failed, skipping", zap.Error(err))

		const errText = "2. Вложение: стикер (AddSticker)\n\n" +
			"Не удалось отправить стикер. Метод msg.AddSticker(code) прикрепляет стикер по его коду. " +
			"Стикер должен быть единственным вложением в сообщении."

		if err := sendWithNav(p, errText); err != nil {
			return nil, fmt.Errorf("send sticker error explanation: %w", err)
		}

		return &State{Scenario: a.Name(), Step: 2}, nil
	}

	// Send explanation and navigation as a separate message.
	if err := sendWithNav(p, text); err != nil {
		return nil, fmt.Errorf("send sticker explanation: %w", err)
	}

	return &State{Scenario: a.Name(), Step: 2}, nil
}

func (a AttachmentsScenario) stepContact(ctx context.Context, p Params) (*State, error) {
	upd := p.Ctx.Update()

	const text = "3. Вложение: контакт (AddContact)\n\n" +
		"Метод msg.AddContact(userID) прикрепляет контакт пользователя по его ID. "

	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		AddContact(upd.UserID)

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("send contact attachment: %w", err)
	}

	// Send explanation and navigation as a separate message.
	if err := sendWithNav(p, text); err != nil {
		return nil, fmt.Errorf("send contact explanation: %w", err)
	}

	return &State{Scenario: a.Name(), Step: 3}, nil
}

func (a AttachmentsScenario) stepLocation(ctx context.Context, p Params) (*State, error) {
	upd := p.Ctx.Update()

	const text = "4. Вложение: локация (AddLocation)\n\n" +
		"Метод msg.AddLocation(lat, lng) прикрепляет точку на карте."

	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		SetText(text).
		AddLocation(55.7558, 37.6173) // Moscow coordinates

	msg.AddKeyboard(navKeyboard())

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("send location attachment: %w", err)
	}

	return &State{Scenario: a.Name(), Step: 4}, nil
}

func (a AttachmentsScenario) stepShare(ctx context.Context, p Params) (*State, error) {
	upd := p.Ctx.Update()

	const text = "5. Вложение: ссылка (AddShare)\n\n" +
		"Метод msg.AddShare(link) прикрепляет расшаренную ссылку к сообщению."

	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		SetText(text).
		AddShare("https://max.ru")

	msg.AddKeyboard(navKeyboard())

	if _, err := p.API.Client().Messages.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("send share attachment: %w", err)
	}

	return &State{Scenario: a.Name(), Step: 5}, nil
}

func (a AttachmentsScenario) stepComplete(p Params) (*State, error) {
	if err := p.Ctx.Send("Демо вложений завершено!"); err != nil {
		return nil, fmt.Errorf("send completion message: %w", err)
	}

	return nil, ErrDone
}

// Ensure AttachmentsScenario implements Scenario interface at compile time.
var _ Scenario = AttachmentsScenario{}

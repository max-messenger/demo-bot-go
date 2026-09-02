package scenario

import (
	"fmt"
	"strings"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
)

// SendStartMenu sends the welcome message with inline keyboard menu.
func SendStartMenu(p Params, scenarios []Scenario, isPersonal bool) error {
	var sb strings.Builder
	sb.WriteString("Привет! Я бот-демонстратор API MAX.\n\n")
	sb.WriteString("Я показываю все возможности библиотеки max-bot-api-client-go. ")
	sb.WriteString("Выберите сценарий из меню ниже или отправьте команду.\n\n")

	if isPersonal {
		sb.WriteString("📌 Личные сценарии:\n")
	} else {
		sb.WriteString("📌 Групповые сценарии:\n")
	}

	for _, s := range scenarios {
		fmt.Fprintf(&sb, "• /%s — %s\n", s.Name(), s.Description())
	}

	keyboard := buildMenuKeyboard(scenarios)

	return p.Ctx.Send(sb.String(), maxbot.WithKeyboard(keyboard))
}

// SendHelpText sends a plain text help message with bold permission names.
func SendHelpText(p Params, scenarios []Scenario) error {
	var sb strings.Builder
	sb.WriteString("Доступные команды:\n\n")

	for _, s := range scenarios {
		fmt.Fprintf(&sb, "/%s — %s\n", s.Name(), s.Description())

		if len(s.RequiredPermissions()) > 0 {
			sb.WriteString(formatPermissions(s.RequiredPermissions()))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n/start — главное меню\n/help — эта справка")

	upd := p.Ctx.Update()

	msg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		SetText(sb.String()).
		SetFormat(model.FormatMarkdown)

	_, err := p.API.Client().Messages.Send(p.Ctx.Context(), msg)

	return err
}

func buildMenuKeyboard(scenarios []Scenario) *model.Keyboard {
	keyboard := model.NewKeyboard()

	for _, s := range scenarios {
		label := fmt.Sprintf("/%s", s.Name())
		keyboard.AddRow().AddCallBack(label, fmt.Sprintf("s:%s", s.Name()))
	}

	return keyboard
}

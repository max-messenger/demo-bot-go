package scenario

import (
	"fmt"
	"strings"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

// SendGroupHelpText sends a help message for group chat context with bold permission names.
func SendGroupHelpText(p Params, scenarios []Scenario) error {
	var sb strings.Builder
	sb.WriteString("Групповые команды:\n\n")

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

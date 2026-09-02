package scenario

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
)

// PayloadCancel is the callback payload for the "Завершить"/"Отмена" button.
const PayloadCancel = "cancel"

// PayloadCmdStart is the callback payload for the /start button.
const PayloadCmdStart = "cmd:start"

// PayloadCmdHelp is the callback payload for the /help button.
const PayloadCmdHelp = "cmd:help"

// navKeyboard builds an inline keyboard with "Далее" and "Завершить" buttons.
func navKeyboard() *model.Keyboard {
	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("Далее", "next")
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	return kb
}

// cancelKeyboard builds an inline keyboard with only a "Завершить" button.
func cancelKeyboard() *model.Keyboard {
	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("Завершить", PayloadCancel)

	return kb
}

// sendWithNav sends a text message with a navigation keyboard (Далее + Завершить).
func sendWithNav(p Params, text string) error {
	return p.Ctx.Send(text, maxbot.WithKeyboard(navKeyboard()))
}

// replyKeyboard builds a reply_keyboard using the Keyboard builder.
func replyKeyboard() *model.Keyboard {
	return model.NewKeyboard()
}

// getMessageID extracts a message ID string from the scenario state data.
func getMessageID(s *State, key string) (string, error) {
	if s.Data == nil {
		return "", fmt.Errorf("state data is nil")
	}

	val, ok := s.Data[key]
	if !ok {
		return "", fmt.Errorf("%s not found in state data", key)
	}

	switch v := val.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unexpected %s type: %T", key, val)
	}
}

// getStateString extracts a string value from the scenario state data.
func getStateString(s *State, key string) (string, error) {
	if s.Data == nil {
		return "", fmt.Errorf("state data is nil")
	}

	val, ok := s.Data[key]
	if !ok {
		return "", fmt.Errorf("%s not found in state data", key)
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("unexpected %s type: %T", key, val)
	}

	return str, nil
}

// getStateInt64 extracts an int64 value from the scenario state data.
func getStateInt64(s *State, key string) (int64, error) {
	if s.Data == nil {
		return 0, fmt.Errorf("state data is nil")
	}

	val, ok := s.Data[key]
	if !ok {
		return 0, fmt.Errorf("%s not found in state data", key)
	}

	switch v := val.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(math.Round(v)), nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse %s string value: %w", key, err)
		}

		return parsed, nil
	default:
		return 0, fmt.Errorf("unexpected %s type: %T", key, val)
	}
}

// parseUserID parses a string as int64 user ID.
func parseUserID(s string) (int64, error) {
	s = strings.TrimSpace(s)

	userID, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse user id %q: %w", s, err)
	}

	return userID, nil
}

// MenuKeyboard builds an inline keyboard with /start and /help buttons.
func MenuKeyboard() *model.Keyboard {
	kb := model.NewKeyboard()
	kb.AddRow().AddCallBack("/start", PayloadCmdStart)
	kb.AddRow().AddCallBack("/help", PayloadCmdHelp)

	return kb
}

// formatPermissions formats a list of permission names as bold markdown.
func formatPermissions(permissions []string) string {
	var sb strings.Builder
	sb.WriteString("  Требуемые права: ")
	for i, perm := range permissions {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "**%s**", perm)
	}

	return sb.String()
}

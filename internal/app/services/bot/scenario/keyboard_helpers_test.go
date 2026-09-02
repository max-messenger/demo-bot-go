package scenario_test

import (
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"demo_bot/internal/app/services/bot/scenario"
)

func TestUnitMenuKeyboard(t *testing.T) {
	t.Parallel()

	kb := scenario.MenuKeyboard()
	require.NotNil(t, kb, "MenuKeyboard should return a non-nil keyboard")

	attach := kb.Build()
	require.Equal(t, model.AttachInlineKeyboard, attach.Type, "keyboard should be inline")

	buttons := attach.Payload.Buttons
	require.Len(t, buttons, 2, "MenuKeyboard should have 2 rows")

	t.Run("first row has start button", func(t *testing.T) {
		t.Parallel()

		row := buttons[0]
		require.Len(t, row, 1, "first row should have 1 button")

		btn := row[0]
		assert.Equal(t, "/start", btn.Text)
		assert.Equal(t, scenario.PayloadCmdStart, btn.Payload)
		assert.Equal(t, model.ButtonCallback, btn.Type)
	})

	t.Run("second row has help button", func(t *testing.T) {
		t.Parallel()

		row := buttons[1]
		require.Len(t, row, 1, "second row should have 1 button")

		btn := row[0]
		assert.Equal(t, "/help", btn.Text)
		assert.Equal(t, scenario.PayloadCmdHelp, btn.Payload)
		assert.Equal(t, model.ButtonCallback, btn.Type)
	})
}

func TestUnitPayloadConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "cancel", scenario.PayloadCancel)
	assert.Equal(t, "cmd:start", scenario.PayloadCmdStart)
	assert.Equal(t, "cmd:help", scenario.PayloadCmdHelp)
}

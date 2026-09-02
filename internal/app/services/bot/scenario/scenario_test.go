package scenario_test

import (
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/assert"

	"demo_bot/internal/app/services/bot/scenario"
)

func TestUnitIsPersonalChat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		update   model.Update
		expected bool
	}{
		{
			name: "dialog message is personal",
			update: model.Update{
				Message: &model.MessageUpdate{
					Recipient: model.Recipient{
						ChatType: model.ChatTypeDialog,
					},
				},
			},
			expected: true,
		},
		{
			name: "group message is not personal",
			update: model.Update{
				Message: &model.MessageUpdate{
					Recipient: model.Recipient{
						ChatType: model.ChatTypeChat,
					},
				},
			},
			expected: false,
		},
		{
			name: "nil message defaults to personal",
			update: model.Update{
				Message: nil,
			},
			expected: true,
		},
		{
			name: "channel message is not personal",
			update: model.Update{
				Message: &model.MessageUpdate{
					Recipient: model.Recipient{
						ChatType: model.ChatTypeChannel,
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := scenario.IsPersonalChat(tt.update)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnitIsGroupChat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		update   model.Update
		expected bool
	}{
		{
			name: "group message is group",
			update: model.Update{
				Message: &model.MessageUpdate{
					Recipient: model.Recipient{
						ChatType: model.ChatTypeChat,
					},
				},
			},
			expected: true,
		},
		{
			name: "dialog message is not group",
			update: model.Update{
				Message: &model.MessageUpdate{
					Recipient: model.Recipient{
						ChatType: model.ChatTypeDialog,
					},
				},
			},
			expected: false,
		},
		{
			name: "nil message defaults to not group",
			update: model.Update{
				Message: nil,
			},
			expected: false,
		},
		{
			name: "channel message is not group",
			update: model.Update{
				Message: &model.MessageUpdate{
					Recipient: model.Recipient{
						ChatType: model.ChatTypeChannel,
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := scenario.IsGroupChat(tt.update)
			assert.Equal(t, tt.expected, result)
		})
	}
}

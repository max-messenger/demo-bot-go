package bigbro

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.uber.org/zap"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestUnitSendEvent(t *testing.T) {
	mockClient := new(MockHTTPClient)
	mockClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
	}, nil)

	config := Config{
		Enabled: true,
		URL:     "http://test.com",
		Token:   "test-token",
	}

	client := Client{
		config: config,
		client: mockClient,
		logger: zap.NewNop(),
	}

	eventFields := map[string]any{
		"key": "value",
	}

	err := client.SendEvent(context.Background(), "test-event", "click", 123, eventFields)
	assert.NoError(t, err)
}

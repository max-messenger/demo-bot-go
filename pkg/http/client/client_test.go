package client_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"demo_bot/pkg/http/client"
)

func TestUnitClient(t *testing.T) {
	suite.Run(t, new(testClient))
}

type testClient struct {
	suite.Suite

	logger *zap.Logger
	client *client.Client
}

func (t *testClient) SetupTest() {
	t.logger = zaptest.NewLogger(t.T())

	before := func(r *http.Request) {
		r.Header.Add("X-Test-Go-Suite", "unit test")
	}

	after := func(r *http.Request, _ *http.Response) {
		t.Assert().Equal("unit test", r.Header.Get("X-Test-Go-Suite"))
	}

	opts := []client.Option{
		client.WithTimeout(time.Second),
		client.WithFuncBefore(before),
		client.WithFuncAfter(after),
		client.WithLogger(t.logger),
	}
	t.client = client.New("test-unit", opts...)
}

func (t *testClient) TestClient() {
	testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	defer testSrv.Close()

	req, err := http.NewRequest(http.MethodGet, testSrv.URL, nil)
	t.NoError(err)

	resp, err := t.client.Do(req)
	t.NoError(err)
	t.Equal(http.StatusAccepted, resp.StatusCode)
}

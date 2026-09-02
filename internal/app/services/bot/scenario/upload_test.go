package scenario_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"demo_bot/internal/app/services/bot/mocks"
	"demo_bot/internal/app/services/bot/scenario"
	scenariomocks "demo_bot/internal/app/services/bot/scenario/mocks"
)

type UploadScenarioSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	api      *scenariomocks.MockAPI
	ctx      *mocks.MockContext
	chats    *mocks.MockChatsAPI
	messages *mocks.MockMessagesAPI
	upload   *mocks.MockUploadAPI
	log      *zap.Logger
}

func (s *UploadScenarioSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.api = scenariomocks.NewMockAPI(s.ctrl)
	s.ctx = mocks.NewMockContext(s.ctrl)
	s.chats = mocks.NewMockChatsAPI(s.ctrl)
	s.messages = mocks.NewMockMessagesAPI(s.ctrl)
	s.upload = mocks.NewMockUploadAPI(s.ctrl)
	s.log = zap.NewNop()

	apiClient := &maxbot.Api{
		Chats:    s.chats,
		Messages: s.messages,
		Upload:   s.upload,
	}
	s.api.EXPECT().Client().Return(apiClient).AnyTimes()
}

func (s *UploadScenarioSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *UploadScenarioSuite) params() scenario.Params {
	s.T().Helper()

	return scenario.Params{
		API: s.api,
		Ctx: s.ctx,
		Log: s.log,
	}
}

func TestUnitUploadScenario(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UploadScenarioSuite))
}

func (s *UploadScenarioSuite) TestStep0_Prompt() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 0},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), "upload", state.Scenario)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *UploadScenarioSuite) TestStep1_NoMessage() {
	s.ctx.EXPECT().Update().Return(model.Update{ChatID: 42}).AnyTimes()

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *UploadScenarioSuite) TestStep1_NoImageAttachment() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{
				Text:        "hello",
				Attachments: []model.Attachment{},
			},
		},
	}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), state)
	assert.Equal(s.T(), 1, state.Step)
}

func (s *UploadScenarioSuite) TestUnknownStep_Error() {
	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 99},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
}

func (s *UploadScenarioSuite) TestMetadata() {
	sc := scenario.NewUploadScenario()

	assert.Equal(s.T(), "upload", sc.Name())
	assert.Equal(s.T(), "демо загрузки и отправки файла", sc.Description())
	assert.Equal(s.T(), scenario.ChatTypePersonal, sc.ChatType())
	assert.Nil(s.T(), sc.RequiredPermissions())
}

func (s *UploadScenarioSuite) TestStep0_SendError() {
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send failed"))

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 0},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
	assert.Contains(s.T(), err.Error(), "send upload prompt")
}

func (s *UploadScenarioSuite) TestStep1_NoImageAttachmentSendError() {
	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{
				Text:        "hello",
				Attachments: []model.Attachment{},
			},
		},
	}).AnyTimes()
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send failed"))

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
	assert.Contains(s.T(), err.Error(), "send no image error")
}

func (s *UploadScenarioSuite) TestStep1_ImageDownloadAndSend() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-image-data"))
	}))
	defer srv.Close()

	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{
				Attachments: []model.Attachment{
					{
						Type: model.AttachImage,
						Payload: model.Payload{URL: srv.URL + "/image.png"},
					},
				},
			},
		},
	}).AnyTimes()
	s.upload.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("token123", nil)
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, scenario.ErrDone)
	assert.Nil(s.T(), state)
}

func (s *UploadScenarioSuite) TestStep1_ImageDownloadNon200Status() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{
				Attachments: []model.Attachment{
					{
						Type: model.AttachImage,
						Payload: model.Payload{URL: srv.URL + "/image.png"},
					},
				},
			},
		},
	}).AnyTimes()

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
	assert.Contains(s.T(), err.Error(), "unexpected status")
}

func (s *UploadScenarioSuite) TestStep1_UploadAPIError() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-image-data"))
	}))
	defer srv.Close()

	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{
				Attachments: []model.Attachment{
					{
						Type: model.AttachImage,
						Payload: model.Payload{URL: srv.URL + "/image.png"},
					},
				},
			},
		},
	}).AnyTimes()
	s.upload.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("upload api error"))

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
	assert.Contains(s.T(), err.Error(), "upload image")
}

func (s *UploadScenarioSuite) TestStep1_SendUploadedImageError() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-image-data"))
	}))
	defer srv.Close()

	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{
				Attachments: []model.Attachment{
					{
						Type: model.AttachImage,
						Payload: model.Payload{URL: srv.URL + "/image.png"},
					},
				},
			},
		},
	}).AnyTimes()
	s.upload.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("token123", nil)
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, errors.New("send failed"))

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
	assert.Contains(s.T(), err.Error(), "send uploaded image")
}

func (s *UploadScenarioSuite) TestStep1_CompletionSendError() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-image-data"))
	}))
	defer srv.Close()

	s.ctx.EXPECT().Update().Return(model.Update{
		ChatID: 42,
		UserID: 10,
		Message: &model.MessageUpdate{
			Body: model.MessageBody{
				Attachments: []model.Attachment{
					{
						Type: model.AttachImage,
						Payload: model.Payload{URL: srv.URL + "/image.png"},
					},
				},
			},
		},
	}).AnyTimes()
	s.upload.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("token123", nil)
	s.messages.EXPECT().Send(gomock.Any(), gomock.Any()).Return(model.SendMessageResult{}, nil)
	s.ctx.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("completion send failed"))

	sc := scenario.NewUploadScenario()
	state, err := sc.Handle(
		context.Background(), s.params(),
		&scenario.State{Scenario: "upload", Step: 1},
	)

	require.Error(s.T(), err)
	assert.Nil(s.T(), state)
	assert.Contains(s.T(), err.Error(), "send completion message")
}

package scenario

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	maxbotcli "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
	maxbot "github.com/max-messenger/maxbot"
)

// httpClient is a shared HTTP client with sane timeouts for downloading images.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// UploadScenario demonstrates file upload via the Upload API.
type UploadScenario struct{}

// NewUploadScenario creates a new upload scenario.
func NewUploadScenario() *UploadScenario {
	return &UploadScenario{}
}

// Name returns the scenario command name.
func (UploadScenario) Name() string { return "upload" }

// Description returns a short Russian description for /help.
func (UploadScenario) Description() string {
	return "демо загрузки и отправки файла"
}

// ChatType returns where this scenario can run.
func (UploadScenario) ChatType() ChatType { return ChatTypePersonal }

// RequiredPermissions returns admin permissions needed (group only).
func (UploadScenario) RequiredPermissions() []string { return nil }

// Handle processes one step of the upload scenario.
func (u UploadScenario) Handle(ctx context.Context, p Params, s *State) (*State, error) {
	switch s.Step {
	case 0:
		return u.step0(p)
	case 1:
		return u.step1(ctx, p)
	default:
		return nil, fmt.Errorf("upload: unknown step %d", s.Step)
	}
}

func (u UploadScenario) step0(p Params) (*State, error) {
	if err := p.Ctx.Send(
		"Отправьте мне изображение, и я загружу и отправлю его обратно",
		maxbot.WithKeyboard(cancelKeyboard()),
	); err != nil {
		return nil, fmt.Errorf("send upload prompt: %w", err)
	}

	return &State{Scenario: u.Name(), Step: 1}, nil
}

func (u UploadScenario) step1(ctx context.Context, p Params) (*State, error) {
	upd := p.Ctx.Update()
	msg := upd.Message
	if msg == nil {
		return &State{Scenario: u.Name(), Step: 1}, nil
	}

	imageURL := extractImageURL(msg.Body.Attachments)
	if imageURL == "" {
		if err := p.Ctx.Send("В сообщении нет изображения. Пожалуйста, отправьте изображение."); err != nil {
			return nil, fmt.Errorf("send no image error: %w", err)
		}

		return &State{Scenario: u.Name(), Step: 1}, nil
	}

	if err := u.downloadAndSendImage(ctx, p, imageURL); err != nil {
		return nil, err
	}

	return nil, ErrDone
}

func (u UploadScenario) downloadAndSendImage(ctx context.Context, p Params, imageURL string) error {
	upd := p.Ctx.Update()

	body, size, err := downloadImage(ctx, imageURL)
	if err != nil {
		return err
	}
	defer body.Close()

	token, err := p.API.Client().Upload.Upload(ctx, model.UploadImage, body, "image.png", size)
	if err != nil {
		return fmt.Errorf("upload image: %w", err)
	}

	resultMsg := maxbotcli.NewMessage().
		SetUser(upd.UserID).
		SetChat(upd.ChatID).
		SetText("Изображение загружено через Upload.Upload() и отправлено через AddAttachByToken():").
		AddAttachByToken(token, model.AttachImage)

	if _, err := p.API.Client().Messages.Send(ctx, resultMsg); err != nil {
		return fmt.Errorf("send uploaded image: %w", err)
	}

	if err := p.Ctx.Send("Демо загрузки завершено!"); err != nil {
		return fmt.Errorf("send completion message: %w", err)
	}

	return nil
}

func downloadImage(ctx context.Context, imageURL string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create download request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("download image: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()

		return nil, 0, fmt.Errorf("download image: unexpected status %d", resp.StatusCode)
	}

	size := resp.ContentLength
	if size <= 0 {
		size = 0
	}

	return resp.Body, size, nil
}

func extractImageURL(attachments []model.Attachment) string {
	for _, att := range attachments {
		if att.Type == model.AttachImage && att.Payload.URL != "" {
			return att.Payload.URL
		}
	}

	return ""
}

// Ensure UploadScenario implements Scenario interface at compile time.
var _ Scenario = UploadScenario{}

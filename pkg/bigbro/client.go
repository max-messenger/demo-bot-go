package bigbro

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"demo_bot/pkg/marshaler"
	"demo_bot/pkg/telemetry"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	logger *zap.Logger
	config Config
	client HTTPClient
}

func New(logger *zap.Logger, config Config, cli HTTPClient) (*Client, error) {
	return &Client{
		client: cli,
		logger: logger,
		config: config,
	}, nil
}

func (c *Client) SendEvent(
	ctx context.Context,
	eventName string,
	eventType EventType,
	userID int64,
	eventFields EventFields,
) error {
	if !c.config.Enabled {
		return nil
	}

	var err error
	defer func(t time.Time) {
		metricRequestsTotal.WithLabelValues("send-event", telemetry.ErrLabelValue(err)).Inc()
		metricRequestsDuration.WithLabelValues("send-event").Observe(time.Since(t).Seconds())
	}(time.Now())

	eventDTO := eventBuilder(eventName, eventType, userID, eventFields, c.config.ExtraFields)

	jsonBody, err := marshaler.MarshalJSON(eventDTO)
	if err != nil {
		return fmt.Errorf("bigbro: MarshalJSON: %w", err)
	}

	endpoint, err := url.JoinPath(c.config.URL, analyticsEventPath)
	if err != nil {
		return fmt.Errorf("bigbro: JoinPath: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("bigbro: NewRequestWithContext: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.config.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("bigbro: client.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bigbro: unexpected status code: %v", resp.StatusCode)
	}

	result := new(eventResponse)
	err = marshaler.LoadJSONFromReader(resp.Body, result)
	if err != nil {
		return fmt.Errorf("bigbro: LoadJSONFromReader: %v", resp.StatusCode)
	}

	c.logger.Debug("bigbro send event",
		zap.String("eventName", eventName),
		zap.Int64("userID", userID),
		zap.Int("status", resp.StatusCode),
		zap.Any("response", result),
	)

	if !result.OK {
		return fmt.Errorf("bigbro: invalid response: %+v", result)
	}

	return nil
}

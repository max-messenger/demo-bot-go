package client

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"demo_bot/pkg/telemetry"
)

const defaultTimeout = 30 * time.Second

type Option func(*Client)

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpCli.Timeout = timeout
	}
}

func WithFuncBefore(f func(*http.Request)) Option {
	return func(c *Client) {
		c.funcBefore = f
	}
}

func WithFuncAfter(f func(*http.Request, *http.Response)) Option {
	return func(c *Client) {
		c.funcAfter = f
	}
}

func WithFuncError(f func(*http.Request, error)) Option {
	return func(c *Client) {
		c.funcError = f
	}
}

func WithTransport(transport *http.Transport) Option {
	return func(c *Client) {
		c.httpCli.Transport = otelhttp.NewTransport(transport)
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(c *Client) {
		c.log = logger
	}
}

type Client struct {
	name    string
	httpCli *http.Client
	log     *zap.Logger

	funcBefore func(*http.Request)
	funcAfter  func(*http.Request, *http.Response)
	funcError  func(*http.Request, error)
}

func New(name string, opts ...Option) *Client {
	cli := &Client{
		name: name,
		httpCli: &http.Client{
			Timeout: defaultTimeout,
			Transport: otelhttp.NewTransport(
				&http.Transport{
					TLSClientConfig: &tls.Config{
						MinVersion: tls.VersionTLS13,
					},
				},
			),
		},
	}

	for _, opt := range opts {
		opt(cli)
	}

	return cli
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var err error
	var resp *http.Response
	var statusCode int

	defer func(t time.Time) {
		metricHTTPRequestsTotal.WithLabelValues(c.name, req.Method, strconv.Itoa(statusCode), telemetry.ErrLabelValue(err)).Inc()
		metricHTTPRequestsDuration.WithLabelValues(c.name, req.Method).Observe(time.Since(t).Seconds())
	}(time.Now())

	c.log.Info(
		"request",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
	)

	if c.funcBefore != nil {
		c.funcBefore(req)
	}
	resp, err = c.httpCli.Do(req) //nolint:gosec

	if c.funcError != nil {
		c.funcError(req, err)
	}

	if err != nil {
		return nil, err
	}

	if c.funcAfter != nil {
		c.funcAfter(req, resp)
	}

	statusCode = resp.StatusCode
	c.log.Info(
		"response",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Int("status", statusCode))

	return resp, nil
}

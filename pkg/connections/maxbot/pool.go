package maxbot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/max-messenger/maxbot"
	"github.com/max-messenger/maxbot/middleware"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

var ErrPoolNotFound = errors.New("pool not found")

type PoolConfig map[string]Config

type Config struct {
	ApiUrl string   `yaml:"api_url"` // nolint:all
	Token  string   `yaml:"token"`
	URL    string   `yaml:"url"`
	Path   string   `yaml:"path"`
	Secret string   `yaml:"secret"` // nolint:gosec
	Types  []string `yaml:"types"`
}

type Pool struct {
	configs PoolConfig
	pools   map[string]*maxbot.Api
	log     *zap.Logger
}

func NewMaxBotPool(params PoolParams, log *zap.Logger) (*Pool, error) {
	p := &Pool{
		configs: params.Configs,
		pools:   make(map[string]*maxbot.Api, len(params.Configs)),
		log:     log,
	}

	for name, cfg := range p.configs {
		if cfg.URL != "" {
			params.Opts = append(params.Opts, maxbot.WithWebhook(cfg.URL+cfg.Path, cfg.Secret, cfg.Types))
		}
		if cfg.ApiUrl != "" {
			params.Opts = append(params.Opts, maxbot.WithBaseURL(cfg.ApiUrl))
		}

		api, err := maxbot.NewApi(cfg.Token, params.Opts...)
		if err != nil {
			return nil, err
		}

		for _, mr := range params.Middlewares {
			api.Use(mr)
		}

		api.Use(middleware.Recover(func(err error, ctx maxbot.Context) {
			log.Error(fmt.Sprintf("message recover: %v", ctx), zap.Error(err))
		}))
		api.Use(func(next maxbot.HandlerFunc) maxbot.HandlerFunc {
			return func(c maxbot.Context) error {
				_, span := otel.Tracer("maxbot").Start(c.Context(), "bot-message")
				defer span.End()

				defer func(t time.Time) {
					metricRequestsTotal.WithLabelValues("bot-message").Inc()
					metricRequestsDuration.WithLabelValues("bot-message").Observe(time.Since(t).Seconds())
				}(time.Now())

				return next(c)
			}
		})
		go api.Start()

		p.pools[name] = api
	}

	return p, nil
}

func (p *Pool) GetPool(name string) (*maxbot.Api, error) {
	if pool, ok := p.pools[name]; ok {
		return pool, nil
	}

	return nil, ErrPoolNotFound
}

func (p *Pool) Register(r chi.Router) {
	mw := webhookHandleMiddleware{}
	for name, cfg := range p.configs {
		r.Handle(cfg.Path, mw.WebhookHandler(p.pools[name].WebhookHandler(), p.log))
	}
}

type webhookHandleMiddleware struct{}

func (w webhookHandleMiddleware) WebhookHandler(h http.HandlerFunc, log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		log.Info("receive webhook", zap.String("path", r.URL.Path), zap.String("body", string(bodyBytes)))
		h(w, r)
	}
}

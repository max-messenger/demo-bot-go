package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HTTPServer struct {
	http.Server

	logger *zap.Logger
}

func NewHTTPServer(
	logger *zap.Logger,
	cfg *HTTPConfig,
	h http.Handler,
) (*HTTPServer, error) {
	if cfg == nil {
		return nil, nil // nolint:nilnil // it's ok
	}
	s := &HTTPServer{
		Server: http.Server{
			Addr:              cfg.Addr,
			Handler:           h,
			ReadHeaderTimeout: time.Second * 3,
		},
		logger: logger,
	}

	return s, nil
}

func (s *HTTPServer) Start(_ context.Context) error {
	ln, err := net.Listen("tcp", s.Addr) // nolint:noctx
	if err != nil {
		return err
	}

	s.logger.Info(
		"Starting HTTP server",
		zap.String("addr", s.Addr),
	)

	go func() {
		if err := s.Serve(ln); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				s.logger.Info("Server shutdown")

				return
			}
			s.logger.Error("Failed to start HTTP server", zap.Error(err))
		}
	}()

	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

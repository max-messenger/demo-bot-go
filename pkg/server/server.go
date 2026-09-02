package server

import (
	"context"

	"go.uber.org/fx"
)

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type ServersParams struct {
	fx.In

	HTTPServer   *HTTPServer
	GRPCServer   *GRPCServer `optional:"true"`
	SystemServer *SystemServer
}
type Servers struct {
	servers []Server
}

func NewServersProvider(params ServersParams) *Servers {
	servers := make([]Server, 0, 3)
	if params.HTTPServer != nil {
		servers = append(servers, params.HTTPServer)
	}

	if params.GRPCServer != nil {
		servers = append(servers, params.GRPCServer)
	}

	servers = append(servers, params.SystemServer)

	return &Servers{servers: servers}

}

func (s *Servers) Start(ctx context.Context) error {
	for _, srv := range s.servers {
		if err := srv.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Servers) Stop(ctx context.Context) error {
	for _, srv := range s.servers {
		if err := srv.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}

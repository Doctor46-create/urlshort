package server

import (
	grpcserver "github.com/Doctor46-create/urlshort/internal/grpc/server"
)

func (a *Application) initGRPCServer() {
	a.grpcServer = grpcserver.NewGRPCServer(
		*a.cfg,
		a.service,
	)
}

package server

import (
	"fmt"
	"net"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/Doctor46-create/urlshort/internal/service"
	pb "github.com/Doctor46-create/urlshort/internal/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	cfg config.Config
	svc service.Shortener
	grpcServer *grpc.Server
}

func NewGRPCServer(cfg config.Config, svc service.Shortener) *GRPCServer {
	return &GRPCServer{
		cfg: cfg,
		svc: svc,
	}
}

func (s *GRPCServer) Start() error {
	listener, err := net.Listen("tcp", s.cfg.GetGRPCAddress())
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	adapter := NewShortenerServerAdapter(s.svc, s.cfg.GetSecretKey())

	s.grpcServer = grpc.NewServer()
	pb.RegisterShortenerServiceServer(s.grpcServer, adapter)

	reflection.Register(s.grpcServer)

	fmt.Printf("gRPC server listening on %s\n", s.cfg.GetGRPCAddress())
	return s.grpcServer.Serve(listener)
}

func (s *GRPCServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

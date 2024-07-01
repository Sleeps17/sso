package grpcserver

import (
	"fmt"
	authgrpc "github.com/sso/internal/grpc/auth"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type Server struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, authService authgrpc.Auth, port int) *Server {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &Server{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (s *Server) MustRun() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		panic("can't create listener: " + err.Error())
	}

	s.log.Info("gRPC server is running: ", slog.String("addr", l.Addr().String()))

	if err := s.gRPCServer.Serve(l); err != nil {
		panic("can't serve gRPC server: " + err.Error())
	}
}

func (s *Server) Stop() {
	s.log.Info("stopping gRPC server")

	s.gRPCServer.GracefulStop()
}

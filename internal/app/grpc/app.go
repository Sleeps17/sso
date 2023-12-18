package grpcapp

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		panic("can't create listener: " + err.Error())
	}

	a.log.Info("gRPC server is running: ", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		panic("can't serve gRPC server: " + err.Error())
	}
}

func (a *App) Stop() {
	a.log.Info("stopping gRPC server")

	a.gRPCServer.GracefulStop()
}

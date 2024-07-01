package app

import (
	"context"
	grpcserver "github.com/sso/internal/app/grpc"
	"github.com/sso/internal/services/auth"
	"github.com/sso/internal/storage/postgresql"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcserver.Server
}

func MustNew(log *slog.Logger, grpcPort int, dbConnString string, dbConnTimeout time.Duration, tokenTTL time.Duration) *App {
	ctx, cancel := context.WithTimeout(context.Background(), dbConnTimeout)
	defer cancel()

	s, err := func() (*postgresql.Storage, error) {
		defer cancel()
		return postgresql.New(ctx, dbConnString)
	}()
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, s, s, tokenTTL)
	grpcServer := grpcserver.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcServer,
	}
}

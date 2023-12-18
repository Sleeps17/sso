package suite

import (
	"context"
	ssov1 "github.com/Sleeps17/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"
)

const grpcHost = "localhost"

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/local.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPCConfig.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	cc, err := grpc.DialContext(ctx, grpcAddres(cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}

func grpcAddres(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.Port))
}

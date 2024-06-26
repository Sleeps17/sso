package main

import (
	"github.com/sso/internal/app"
	"github.com/sso/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// TODO: Завернуть объект storage в отдельное приложение, которое будет иметь методы Run и Stop
// TODO: Подумать над ручкой, которая делает user-a админом
func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("start application",
		slog.String("env", cfg.Env),
	)

	application := app.MustNew(
		log,
		cfg.Server.Port,
		cfg.Storage.Url(),
		cfg.Storage.Timeout,
		cfg.TokenTTL,
	)
	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sig := <-stop
	log.Info("stopping application", slog.String("signal", sig.String()))

	application.GRPCSrv.Stop()

	log.Info("application stopped")
}

const localEnv = "local"
const devEnv = "dev"
const prodEnv = "prod"

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case localEnv:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case devEnv:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case prodEnv:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}),
		)
	}

	return log
}

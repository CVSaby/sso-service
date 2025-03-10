package main

import (
	"context"
	"github.com/CVSaby/sso-service/internal/app"
	"github.com/CVSaby/sso-service/internal/config"
	"github.com/CVSaby/sso-service/internal/lib/otel"
	otlp "go.opentelemetry.io/otel"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()

	log, logProvider, meterProvider := otel.SetupOtel(
		ctx, cfg.ServiceName,
		cfg.ServiceVersion,
		cfg.OTLPEndpoint,
		cfg.Env,
	)

	otlp.SetMeterProvider(meterProvider)

	log.Info("starting application")

	application := app.New(log, cfg.GRPC.Port, cfg.DBConfig, cfg.JWT)
	go application.GRPCApp.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	application.GRPCApp.Stop()
	logProvider.Shutdown(ctx)
	log.Info("application stopped")
}

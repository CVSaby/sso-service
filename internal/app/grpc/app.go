package grpcapp

import (
	"fmt"
	authgrpc "github.com/CVSaby/sso-service/internal/transport/grpc/auth"
	"github.com/CVSaby/sso-service/pkg/interceptors"
	otlp "go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

var tracer = otlp.Tracer("github.com/CVSaby/sso-service")

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, port int, authService authgrpc.Auther) *App {
	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.TracingInterceptor(tracer)),
	)

	authgrpc.Register(gRPCServer, authService)

	return &App{log, gRPCServer, port}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("operation", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("address", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("operation", op)).
		Info("stopping grpc server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}

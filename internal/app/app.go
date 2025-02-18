package app

import (
	grpcapp "github.com/CVSaby/sso-service/internal/app/grpc"
	"log/slog"
	"time"
)

type App struct {
	GRPCApp *grpcapp.App
}

func New(
	logger *slog.Logger,
	grpcPort int,
	tokenTTL time.Duration,
) *App {
	// TODO: init storage

	// TODO: init auth service

	grpcApp := grpcapp.New(logger, grpcPort)

	return &App{
		GRPCApp: grpcApp,
	}
}

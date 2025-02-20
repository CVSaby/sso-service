package app

import (
	grpcapp "github.com/CVSaby/sso-service/internal/app/grpc"
	"github.com/CVSaby/sso-service/internal/config"
	"github.com/CVSaby/sso-service/internal/services/auth"
	"github.com/CVSaby/sso-service/internal/storage/psql"
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
	dbConfig config.DBConfig,
	jwtCfg config.JWTConfig,
) *App {
	store, err := psql.New(dbConfig)
	if err != nil {
		panic(err)
	}

	authSvc := auth.New(logger, store, store, jwtCfg)

	grpcApp := grpcapp.New(logger, grpcPort, authSvc)

	return &App{
		GRPCApp: grpcApp,
	}
}

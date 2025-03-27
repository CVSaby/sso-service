package app

import (
	grpcapp "github.com/CVSaby/sso-service/internal/app/grpc"
	"github.com/CVSaby/sso-service/internal/config"
	"github.com/CVSaby/sso-service/internal/services/auth"
	"github.com/CVSaby/sso-service/internal/storage/psql"
	kfktransport "github.com/CVSaby/sso-service/internal/transport/kafka"
	"log/slog"
)

type App struct {
	GRPCApp  *grpcapp.App
	KafkaApp *kfktransport.Producer
}

func New(
	logger *slog.Logger,
	grpcPort int,
	dbConfig config.DBConfig,
	jwtCfg config.JWTConfig,
	kfkConfig config.KafkaConfig,
) *App {
	store, err := psql.New(dbConfig)
	if err != nil {
		panic(err)
	}

	kfkProducer := kfktransport.New(kfkConfig.Servers, kfkConfig.ClientID, kfkConfig.Topic)

	authSvc := auth.New(logger, store, store, kfkProducer, jwtCfg)

	grpcApp := grpcapp.New(logger, grpcPort, authSvc)

	return &App{
		GRPCApp:  grpcApp,
		KafkaApp: kfkProducer,
	}
}

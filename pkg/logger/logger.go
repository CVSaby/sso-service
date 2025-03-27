package logger

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"log/slog"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func SetupLogger(
	env string,
	res *resource.Resource,
	exporter *otlploghttp.Exporter,
) (*slog.Logger, *log.LoggerProvider) {

	switch env {
	case envLocal:
		return consoleLogger()
	case envProd:
		return httpLogger(exporter, res)
	default:
		return httpLogger(exporter, res)
	}
}

func consoleLogger() (*slog.Logger, *log.LoggerProvider) {
	exp, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
	if err != nil {
		panic(err)
	}

	provider := log.NewLoggerProvider(
		log.WithProcessor(log.NewSimpleProcessor(exp)),
	)

	logger := otelslog.NewLogger("sso-service-logger", otelslog.WithLoggerProvider(provider))

	return logger, provider
}

func httpLogger(exp *otlploghttp.Exporter, res *resource.Resource) (*slog.Logger, *log.LoggerProvider) {
	logProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exp)),
	)

	logger := otelslog.NewLogger("sso-service-logger", otelslog.WithSource(true), otelslog.WithLoggerProvider(logProvider))

	return logger, logProvider
}

package otel

import (
	"context"
	"github.com/CVSaby/sso-service/internal/lib/logger"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"log/slog"
	"time"
)

func SetupOtel(ctx context.Context, serviceName string, serviceVersion string, OTLPEndpoint string, env string) (*slog.Logger, *log.LoggerProvider, *metric.MeterProvider) {
	// OTLP Resource
	res, err := NewResource(serviceName, serviceVersion)
	if err != nil {
		panic(err)
	}

	logger, logProvider := setupLogger(ctx, env, res, OTLPEndpoint)
	meterProvider := setupMeter(ctx, res, OTLPEndpoint)

	return logger, logProvider, meterProvider
}

func setupMeter(ctx context.Context, res *resource.Resource, OTLPEndpoint string) *metric.MeterProvider {
	meterExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(OTLPEndpoint), otlpmetrichttp.WithInsecure())
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(meterExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider
}

func setupLogger(ctx context.Context, env string, res *resource.Resource, otlpEndpoint string) (*slog.Logger, *log.LoggerProvider) {
	// logs
	exp, err := otlploghttp.New(ctx, otlploghttp.WithInsecure(), otlploghttp.WithEndpoint(otlpEndpoint))
	if err != nil {
		panic(err)
	}

	return logger.SetupLogger(env, res, exp)
}

func NewResource(serviceName string, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
}

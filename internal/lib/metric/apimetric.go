package appmetrics

import (
	otlp "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otlp.Meter("github.com/CVSaby/sso-service/")

var ApiCounter, _ = meter.Int64Counter(
	"grpc_api_total_calls",
	metric.WithDescription("Numbers of API calls."),
	metric.WithUnit("{{call}}"),
)

var ApiResponseTime, _ = meter.Float64Histogram(
	"grpc_api_latency",
	metric.WithDescription("The latency of api method"),
	metric.WithUnit("s"),
)

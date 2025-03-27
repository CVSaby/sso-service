package appmetrics

import (
	otlp "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otlp.Meter("github.com/CVSaby/sso-service/")

var RequestsCounter, _ = meter.Int64Counter(
	"sso_grpc_requests_total",
	metric.WithDescription("Total number of HTTP requests"),
	metric.WithUnit("{{call}}"),
)

var RequestLatency, _ = meter.Float64Histogram(
	"sso_grpc_request_duration_seconds",
	metric.WithDescription("Duration of HTTP requests"),
	metric.WithUnit("s"),
)

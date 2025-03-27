package interceptors

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TracingInterceptor fetches ```x-trace-id``` from metadata and creates a span with existing trace id, else creates new
func TracingInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var span trace.Span

		// Get trace ID from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			traceIDs := md.Get("x-trace-id")
			if len(traceIDs) > 0 {
				traceIDString := traceIDs[0]
				traceID, err := trace.TraceIDFromHex(traceIDString)
				if err == nil {
					spanContext := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID: traceID,
					})
					ctx = trace.ContextWithSpanContext(ctx, spanContext)
				}
			}
		}

		// Start a new span
		ctx, span = tracer.Start(ctx,
			fmt.Sprintf("%s %s", "grpcServer", info.FullMethod),
			trace.WithSpanKind(trace.SpanKindInternal),
		)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.system", "grpc"),
		)

		// Continue processing the request
		return handler(ctx, req)
	}
}

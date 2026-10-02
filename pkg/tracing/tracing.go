package tracing

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer initializes OpenTelemetry and returns a clean up function.
func InitTracer(serviceName string) (func(context.Context) error, error) {
	ctx := context.Background()

	// 1. Get Jaeger endpoint from environment
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		// If no endpoint is configured, disable tracing so it doesn't spam logs
		return func(context.Context) error { return nil }, nil
	}

	// Fast connection check to avoid log spam if Jaeger host/DNS is unavailable
	conn, err := net.DialTimeout("tcp", endpoint, 2*time.Second)
	if err != nil {
		// Jaeger is unreachable or DNS lookup failed, safely disable tracing
		return func(context.Context) error { return nil }, nil
	}
	_ = conn.Close()

	// 2. Set up the OTLP exporter over gRPC
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(), // For local Jaeger
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// 3. Define the service resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 4. Create the TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// 5. Register global TracerProvider
	otel.SetTracerProvider(tp)

	// 6. Set global propagator (B3 or TraceContext) to propagate trace IDs across HTTP requests
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Return a cleanup function
	return tp.Shutdown, nil
}

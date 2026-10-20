package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// InitTracer initializes an OpenTelemetry tracer provider with an OTLP/gRPC exporter.
func InitTracer(ctx context.Context, serviceName, endpoint string) (*sdktrace.TracerProvider, error) {
	// Production-grade TLS configuration
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Create the OTLP exporter using secure credentials
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Define the resource with production attributes
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "0.1.0"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentKey.String(env),
			semconv.ServiceVersionKey.String(version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create the Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return tp, nil
}

// SetupGlobalTelemetry configures the global propagator for W3C Trace Context.
func SetupGlobalTelemetry() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

// GetShutdownTimeout returns the configured shutdown timeout for the TracerProvider.
func GetShutdownTimeout() time.Duration {
	timeoutStr := os.Getenv("OTEL_SHUTDOWN_TIMEOUT")
	if timeoutStr == "" {
		return 5 * time.Second
	}
	d, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

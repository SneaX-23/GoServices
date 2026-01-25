package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	Enabled        bool
}

// LoadTelemetryConfig() loads telemetry config from env varibles
func LoadTelemetryConfig() *TelemetryConfig {
	return &TelemetryConfig{
		ServiceName:    getEnv("OTEL_SERVICE_NAME", "auth-service"),
		ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    getEnv("APP_ENV", "development"),
		OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		Enabled:        getEnv("OTEL_ENABLED", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func InitTracer(cfg *TelemetryConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		slog.Info("OpenTlemetry is disabled")
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// OTLP http provider
	ctx := context.Background()
	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // prod WithTLSCientConfig()
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create a tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	slog.Info("OpenTelemetry tracer initialized",
		"service", cfg.ServiceName,
		"endpoint", cfg.OTLPEndpoint,
	)

	// Return shutdown function
	return tp.Shutdown, nil
}

func GetTracer() trace.Tracer {
	return otel.Tracer("github.com/SneaX-23/GoServices/auth-service")
}

// Helper function to start a new span
func StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, spanName)
}

// Records an error in the current span
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
	}
}

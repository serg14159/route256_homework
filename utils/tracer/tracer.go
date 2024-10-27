package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracer initializes tracer provider.
func InitTracer(ctx context.Context, serviceName string, endpointURL string) (*oteltrace.TracerProvider, error) {
	// Initialize OTLP exporter over HTTP
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpointURL),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource describing the service
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment("development"),
			semconv.URLFull(endpointURL),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider
	tp := oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(exp),
		oteltrace.WithResource(r),
	)

	// Set the global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

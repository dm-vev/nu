package langfuse

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"nu/internal/config"
)

// OTELTracer implements tracing using OpenTelemetry sending to Langfuse
type OTELTracer struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
	exporter       *otlptrace.Exporter
	enabled        bool
	config         Config
	IncludeContent bool
}

// NewOTELTracer creates a new OTEL-based Langfuse tracer
func NewOTELTracer(customConfig ...Config) (*OTELTracer, error) {
	// Get global configuration
	cfg := config.Get()

	// Use custom config if provided, otherwise use global config
	var tracerConfig Config
	if len(customConfig) > 0 {
		tracerConfig = customConfig[0]
	} else {
		tracerConfig = Config{
			Enabled:        cfg.Tracing.Langfuse.Enabled,
			SecretKey:      cfg.Tracing.Langfuse.SecretKey,
			PublicKey:      cfg.Tracing.Langfuse.PublicKey,
			Host:           cfg.Tracing.Langfuse.Host,
			Environment:    cfg.Tracing.Langfuse.Environment,
			IncludeContent: cfg.Tracing.Langfuse.IncludeContent,
		}
	}

	if !tracerConfig.Enabled {
		return &OTELTracer{
			enabled: false,
		}, nil
	}

	// Validate required configuration
	if tracerConfig.SecretKey == "" || tracerConfig.PublicKey == "" {
		return nil, fmt.Errorf("langfuse secret key and public key are required")
	}

	if tracerConfig.Host == "" {
		tracerConfig.Host = "https://cloud.langfuse.com"
	}

	// Build Basic Auth header for Langfuse
	auth := base64.StdEncoding.EncodeToString([]byte(tracerConfig.PublicKey + ":" + tracerConfig.SecretKey))

	// Create OTLP HTTP exporter pointing to Langfuse
	ctx := context.Background()

	// Configure endpoint URL properly
	endpointURL := tracerConfig.Host + "/api/public/otel/v1/traces"

	exporterOptions := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(endpointURL),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + auth,
		}),
	}

	// Only use insecure if explicitly using HTTP
	if len(tracerConfig.Host) >= 7 && tracerConfig.Host[:7] == "http://" {
		exporterOptions = append(exporterOptions, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, exporterOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("agent-sdk-go"),
			semconv.ServiceVersionKey.String("1.0.0"),
			attribute.String("langfuse.environment", tracerConfig.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set as global tracer provider
	otel.SetTracerProvider(tp)

	// Create tracer
	tracer := tp.Tracer("agent-sdk-go")

	return &OTELTracer{
		tracerProvider: tp,
		tracer:         tracer,
		exporter:       exporter,
		enabled:        true,
		config:         tracerConfig,
		IncludeContent: tracerConfig.IncludeContent,
	}, nil
}

// ShouldIncludeContent returns whether actual prompt/response content should be included in traces
func (t *OTELTracer) ShouldIncludeContent() bool {
	return t.IncludeContent
}

// Flush flushes the OTEL tracer provider
func (t *OTELTracer) Flush() error {
	if !t.enabled || t.tracerProvider == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return t.tracerProvider.ForceFlush(ctx)
}

// Shutdown shuts down the tracer provider
func (t *OTELTracer) Shutdown() error {
	if !t.enabled || t.tracerProvider == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return t.tracerProvider.Shutdown(ctx)
}

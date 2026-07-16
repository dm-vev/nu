package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// Tracer implements tracing using OpenTelemetry
type Tracer struct {
	tracer      trace.Tracer
	enabled     bool
	serviceName string
}

// Span wraps an OpenTelemetry span to implement contracts.Span
type Span struct {
	span trace.Span
}

// End implements contracts.Span
func (s *Span) End() {
	s.span.End()
}

// AddEvent implements contracts.Span
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", v)))
	}
	s.span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttribute implements contracts.Span
func (s *Span) SetAttribute(key string, value interface{}) {
	s.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}

func (s *Span) RecordError(err error) {
	s.span.RecordError(err)
}

// Config contains configuration for OpenTelemetry
type Config struct {
	// Enabled determines whether OpenTelemetry tracing is enabled
	Enabled bool

	// ServiceName is the name of the service
	ServiceName string

	// CollectorEndpoint is the endpoint of the OpenTelemetry collector
	CollectorEndpoint string

	// Tracer allows passing a pre-built tracer instead of creating one
	Tracer trace.Tracer
}

// NewTracer creates a new OpenTelemetry tracer
func NewTracer(config Config) (*Tracer, error) {
	if !config.Enabled {
		return &Tracer{
			enabled: false,
		}, nil
	}

	var tracer trace.Tracer

	// Use provided tracer or create a new one
	if config.Tracer != nil {
		tracer = config.Tracer
	} else {
		// Create exporter
		ctx := context.Background()
		exporter, err := otlptrace.New(
			ctx,
			otlptracegrpc.NewClient(
				otlptracegrpc.WithEndpoint(config.CollectorEndpoint),
				otlptracegrpc.WithInsecure(),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}

		// Create resource
		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceNameKey.String(config.ServiceName),
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
		otel.SetTracerProvider(tp)

		// Create tracer
		tracer = tp.Tracer(config.ServiceName)
	}

	return &Tracer{
		tracer:      tracer,
		enabled:     true,
		serviceName: config.ServiceName,
	}, nil
}

// NewTracerWrapper creates a new OpenTelemetry tracer wrapper from existing tracer
func NewTracerWrapper(tracer trace.Tracer) *Tracer {
	if tracer == nil {
		return &Tracer{
			enabled: false,
		}
	}

	return &Tracer{
		tracer:  tracer,
		enabled: true,
	}
}

// StartSpan implements contracts.Tracer
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, contracts.Span) {
	if !t.enabled {
		return ctx, &Span{span: trace.SpanFromContext(ctx)}
	}

	// Get organization ID from context
	orgID, _ := multitenancy.GetOrgID(ctx)

	attrs := []attribute.KeyValue{}
	if orgID != "" {
		attrs = append(attrs, attribute.String("org_id", orgID))
	}

	// Namespace the span name with the library name
	namespacedName := "nu/" + name

	// Start span
	ctx, span := t.tracer.Start(ctx, namespacedName, trace.WithAttributes(attrs...))
	return ctx, &Span{span: span}
}

// StartTraceSession implements contracts.Tracer
func (t *Tracer) StartTraceSession(ctx context.Context, contextID string) (context.Context, contracts.Span) {
	if !t.enabled {
		return ctx, &Span{span: trace.SpanFromContext(ctx)}
	}

	// Get organization ID from context
	orgID, _ := multitenancy.GetOrgID(ctx)

	attrs := []attribute.KeyValue{
		attribute.String("trace.session_id", contextID),
	}
	if orgID != "" {
		attrs = append(attrs, attribute.String("org_id", orgID))
	}

	// Namespace the span name with the library name
	namespacedName := "nu/trace-session"

	// Start root span for the session
	ctx, span := t.tracer.Start(ctx, namespacedName, trace.WithAttributes(attrs...))
	return ctx, &Span{span: span}
}

// @deprecated Use NewTracedLLM - removing in v1.0.0
func NewMemoryOTelMiddleware(memory contracts.Memory, tracer *Tracer) contracts.Memory {
	return NewTracedMemory(memory, tracer)
}

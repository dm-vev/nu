package langfuse

import (
	"context"
	"time"

	"github.com/dm-vev/nu/contracts"
)

// TracerAdapter adapts OTELTracer to implement contracts.Tracer
// This allows the OTEL-based Langfuse tracer to be used with Agents
type TracerAdapter struct {
	otelTracer *OTELTracer
}

// TraceGeneration forwards generation tracing to the wrapped Langfuse tracer.
func (a *TracerAdapter) TraceGeneration(ctx context.Context, modelName, prompt, response string, startTime, endTime time.Time, metadata map[string]interface{}) (string, error) {
	return a.otelTracer.TraceGeneration(ctx, modelName, prompt, response, startTime, endTime, metadata)
}

// NewTracerAdapter creates a new adapter for OTELTracer
func NewTracerAdapter(otelTracer *OTELTracer) contracts.Tracer {
	return &TracerAdapter{
		otelTracer: otelTracer,
	}
}

// StartSpan implements contracts.Tracer by delegating to OTELTracer
func (a *TracerAdapter) StartSpan(ctx context.Context, name string) (context.Context, contracts.Span) {
	return a.otelTracer.StartSpan(ctx, name)
}

// StartTraceSession implements contracts.Tracer by delegating to OTELTracer
func (a *TracerAdapter) StartTraceSession(ctx context.Context, contextID string) (context.Context, contracts.Span) {
	return a.otelTracer.StartTraceSession(ctx, contextID)
}

// Helper function to create and return the adapter in one call
// This makes it easy to migrate existing code
func NewOTELTracerAsInterface(customConfig ...Config) (contracts.Tracer, error) {
	otelTracer, err := NewOTELTracer(customConfig...)
	if err != nil {
		return nil, err
	}

	return NewTracerAdapter(otelTracer), nil
}

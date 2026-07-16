package langfuse

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OTELSpan wraps an OTEL span to implement the contracts.Span interface
type OTELSpan struct {
	span trace.Span
}

// End implements contracts.Span
func (s *OTELSpan) End() {
	s.span.End()
}

// AddEvent implements contracts.Span
func (s *OTELSpan) AddEvent(name string, attributes map[string]interface{}) {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", v)))
	}
	s.span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttribute implements contracts.Span
func (s *OTELSpan) SetAttribute(key string, value interface{}) {
	s.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}

func (s *OTELSpan) RecordError(err error) {
	s.span.RecordError(err)
}

// StartSpan implements contracts.Tracer
func (t *OTELTracer) StartSpan(ctx context.Context, name string) (context.Context, contracts.Span) {
	if !t.enabled {
		// Return a no-op span if tracing is disabled
		return ctx, &OTELSpan{span: trace.SpanFromContext(ctx)}
	}

	// Get organization ID from context if available
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Get agent name from context if available
	agentName, _ := telemetry.GetAgentName(ctx)

	// Create attributes using proper Langfuse namespace
	attrs := []attribute.KeyValue{
		// Trace-level attributes (for list view)
		attribute.String("langfuse.trace.name", telemetry.GetTraceNameOrDefault(ctx, name)),

		// Observation-level attributes (for detailed view)
		attribute.String("langfuse.environment", t.config.Environment),
	}
	if orgID != "" {
		attrs = append(attrs, attribute.String("langfuse.user.id", orgID))
	}

	// Add session ID from conversation context if available
	if conversationID, ok := telemetry.GetConversationID(ctx); ok && conversationID != "" {
		attrs = append(attrs, attribute.String("langfuse.session.id", conversationID))
	}

	// Add agent name if available
	if agentName != "" {
		// Use the correct Langfuse observation metadata format
		attrs = append(attrs, attribute.String("langfuse.observation.metadata.agent_name", agentName))
		// Also try as trace metadata
		attrs = append(attrs, attribute.String("langfuse.trace.metadata.agent_name", agentName))
		// Standard service name (common in observability)
		attrs = append(attrs, attribute.String("service.name", agentName))
		// User-friendly name
		attrs = append(attrs, attribute.String("component.name", agentName))
	}

	// Start OTEL span
	ctx, span := t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))

	// Return wrapped span
	return ctx, &OTELSpan{span: span}
}

// TraceSpan traces a span of execution
func (t *OTELTracer) TraceSpan(ctx context.Context, name string, startTime time.Time, endTime time.Time, metadata map[string]interface{}, parentID string) (string, error) {
	if !t.enabled {
		return "", nil
	}

	// Get organization ID from context
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Get agent name from context if available
	agentName, _ := telemetry.GetAgentName(ctx)

	// Create span
	_, span := t.tracer.Start(ctx, name,
		trace.WithTimestamp(startTime),
		trace.WithAttributes(
			// Trace-level attributes (for list view)
			attribute.String("langfuse.trace.name", telemetry.GetTraceNameOrDefault(ctx, name)),

			// Observation-level attributes (for detailed view)
			attribute.String("langfuse.environment", t.config.Environment),
			attribute.String("langfuse.user.id", orgID),
		),
	)
	defer span.End(trace.WithTimestamp(endTime))

	// Add agent name if available
	if agentName != "" {
		span.SetAttributes(attribute.String("langfuse.observation.metadata.agent_name", agentName))
	}

	// Add metadata as span attributes
	for k, v := range metadata {
		span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", v)))
	}

	return span.SpanContext().SpanID().String(), nil
}

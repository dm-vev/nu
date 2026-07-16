package langfuse

import (
	"context"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StartTraceSession starts a root trace session for a request with the given contextID/requestID
// This creates a root span that will group all subsequent LLM calls and operations
func (t *OTELTracer) StartTraceSession(ctx context.Context, contextID string) (context.Context, contracts.Span) {
	if !t.enabled {
		// Return a no-op span if tracing is disabled
		return ctx, &OTELSpan{span: trace.SpanFromContext(ctx)}
	}

	// Get organization ID from context if available
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Get agent name from context if available
	agentName, _ := telemetry.GetAgentName(ctx)

	// Create root span for the entire request session
	attrs := []attribute.KeyValue{
		// Trace-level attributes (for list view)
		attribute.String("langfuse.trace.name", contextID),

		// Observation-level attributes (for detailed view)
		attribute.String("langfuse.environment", t.config.Environment),
		attribute.String("langfuse.observation.type", "span"),
	}

	if orgID != "" {
		attrs = append(attrs, attribute.String("langfuse.user.id", orgID))
	}

	// Add agent name if available
	if agentName != "" {
		attrs = append(attrs, attribute.String("langfuse.observation.metadata.agent_name", agentName))
	}

	// Add session ID from conversation context if available
	if conversationID, ok := telemetry.GetConversationID(ctx); ok && conversationID != "" {
		attrs = append(attrs, attribute.String("langfuse.session.id", conversationID))
	}

	// Start root OTEL span for the session
	ctx, span := t.tracer.Start(ctx, "request-session", trace.WithAttributes(attrs...))

	// Add contextID to the context for subsequent spans
	ctx = telemetry.WithTraceName(ctx, contextID)
	ctx = telemetry.WithRequestID(ctx, contextID)

	// Return wrapped span
	return ctx, &OTELSpan{span: span}
}

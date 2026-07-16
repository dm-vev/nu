package langfuse

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TraceEvent traces an event
func (t *OTELTracer) TraceEvent(ctx context.Context, name string, input interface{}, output interface{}, level string, metadata map[string]interface{}, parentID string) (string, error) {
	if !t.enabled {
		return "", nil
	}

	// Get organization ID from context
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Get agent name from context if available
	agentName, _ := telemetry.GetAgentName(ctx)

	// Create span for the event
	_, span := t.tracer.Start(ctx, name,
		trace.WithAttributes(
			// Trace-level attributes (for list view)
			attribute.String("langfuse.trace.name", telemetry.GetTraceNameOrDefault(ctx, name)),

			// Observation-level attributes (for detailed view)
			attribute.String("langfuse.observation.level", level),
			attribute.String("langfuse.environment", t.config.Environment),
			attribute.String("langfuse.user.id", orgID),
		),
	)
	defer span.End()

	// Add agent name if available
	if agentName != "" {
		span.SetAttributes(attribute.String("langfuse.observation.metadata.agent_name", agentName))
	}

	// Add trace-level input/output if provided
	if input != nil {
		inputStr := fmt.Sprintf("%v", input)
		span.SetAttributes(
			attribute.String("langfuse.trace.input", inputStr),
			attribute.String("langfuse.observation.input", inputStr),
		)
		span.AddEvent("input", trace.WithAttributes(
			attribute.String("content", inputStr),
		))
	}

	if output != nil {
		outputStr := fmt.Sprintf("%v", output)
		span.SetAttributes(
			attribute.String("langfuse.trace.output", outputStr),
			attribute.String("langfuse.observation.output", outputStr),
		)
		span.AddEvent("output", trace.WithAttributes(
			attribute.String("content", outputStr),
		))
	}

	// Add metadata as span attributes
	for k, v := range metadata {
		span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", v)))
	}

	return span.SpanContext().SpanID().String(), nil
}

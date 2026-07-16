package langfuse

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// createToolCallSpansAsTraceItems creates individual spans for each tool call at the trace root level
func (t *OTELTracer) createToolCallSpansAsTraceItems(ctx context.Context, toolCalls []telemetry.ToolCall) {
	if !t.enabled || len(toolCalls) == 0 {
		fmt.Printf("DEBUG: Tool call spans not created - enabled: %v, toolCalls count: %d\n", t.enabled, len(toolCalls))
		return
	}

	fmt.Printf("DEBUG: Creating %d tool call spans\n", len(toolCalls))

	// Get organization ID from context
	orgID, _ := multitenancy.GetOrgID(ctx)

	for _, toolCall := range toolCalls {
		// Use actual start time if available, otherwise parse timestamp
		var startTime time.Time
		if !toolCall.StartTime.IsZero() {
			startTime = toolCall.StartTime
		} else if toolCall.Timestamp != "" {
			if parsed, err := time.Parse(time.RFC3339, toolCall.Timestamp); err == nil {
				startTime = parsed
			} else {
				startTime = time.Now().Add(-time.Second) // Fallback to 1 second ago
			}
		} else {
			startTime = time.Now().Add(-time.Second)
		}

		// Use actual duration if available, otherwise estimate
		var endTime time.Time
		if toolCall.Duration > 0 {
			endTime = startTime.Add(toolCall.Duration)
		} else {
			endTime = startTime.Add(500 * time.Millisecond) // Default 500ms execution time
		}

		// Create span name that will appear in timeline
		spanName := toolCall.Name

		// Build attributes for the tool call span to appear as a separate timeline item
		attrs := []attribute.KeyValue{
			// Langfuse-specific trace-level attributes (for list view)
			attribute.String("langfuse.trace.name", telemetry.GetTraceNameOrDefault(ctx, spanName)),
			attribute.String("langfuse.trace.input", toolCall.Arguments),
			attribute.String("langfuse.trace.output", toolCall.Result),

			// Langfuse-specific observation-level attributes (for detailed view)
			attribute.String("langfuse.environment", t.config.Environment),
			attribute.String("langfuse.observation.type", "span"),
			attribute.String("langfuse.observation.name", toolCall.Name),
			attribute.String("langfuse.observation.input", toolCall.Arguments),
			attribute.String("langfuse.observation.output", toolCall.Result),

			// Tool-specific attributes
			attribute.String("tool.name", toolCall.Name),
			attribute.String("tool.arguments", toolCall.Arguments),
			attribute.String("tool.result", toolCall.Result),
			attribute.Int64("tool.duration_ms", endTime.Sub(startTime).Milliseconds()),
		}

		// Add organization ID if available
		if orgID != "" {
			attrs = append(attrs, attribute.String("langfuse.user.id", orgID))
		}

		// Add tool call ID if available
		if toolCall.ID != "" {
			attrs = append(attrs, attribute.String("tool.call_id", toolCall.ID))
		}

		// Add error information if present
		if toolCall.Error != "" {
			attrs = append(attrs,
				attribute.String("tool.error", toolCall.Error),
				attribute.String("langfuse.observation.level", "error"),
				attribute.String("langfuse.trace.output", fmt.Sprintf("Error: %s", toolCall.Error)),
				attribute.String("langfuse.observation.output", fmt.Sprintf("Error: %s", toolCall.Error)),
			)
		} else {
			attrs = append(attrs, attribute.String("langfuse.observation.level", "info"))
		}

		// Create tool call span using original context for now
		// Note: These will appear as child spans under LLM generation, but they will be visible
		_, span := t.tracer.Start(ctx, spanName,
			trace.WithTimestamp(startTime),
			trace.WithAttributes(attrs...),
		)

		// Record error if present
		if toolCall.Error != "" {
			span.RecordError(fmt.Errorf("tool execution error: %s", toolCall.Error))
		}

		// End the span with the calculated end time
		span.End(trace.WithTimestamp(endTime))
	}
}

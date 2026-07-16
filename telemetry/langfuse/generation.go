package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// promptToAttributes converts a prompt string to GenAI semantic convention attributes
func (t *OTELTracer) promptToAttributes(prompt string) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	// For simple string prompts, assume it's a user message
	// In the future, this could be enhanced to parse structured prompts
	attrs = append(attrs,
		attribute.String("gen_ai.prompt.0.role", "user"),
		attribute.String("gen_ai.prompt.0.content", prompt),
	)

	return attrs
}

// responseToAttributes converts a response string to GenAI semantic convention attributes
func (t *OTELTracer) responseToAttributes(response string) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	attrs = append(attrs,
		attribute.String("gen_ai.completion.0.role", "assistant"),
		attribute.String("gen_ai.completion.0.content", response),
	)

	return attrs
}

// extractLastUserMessage extracts the last user message from a formatted conversation string
func extractLastUserMessage(conversationText string) string {
	// Handle empty or whitespace-only input
	if strings.TrimSpace(conversationText) == "" {
		return ""
	}

	lines := strings.Split(conversationText, "\n")

	// Look for the last line that starts with "user:"
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "user:") {
			// Remove the "user:" prefix and return the content
			userMessage := strings.TrimSpace(strings.TrimPrefix(line, "user:"))
			if userMessage != "" {
				return userMessage
			}
		}
	}

	// If no user message found, check if the entire text is a single user message
	// (this handles cases where the prompt is just the user input without formatting)
	trimmedText := strings.TrimSpace(conversationText)
	if trimmedText != "" {
		// If the text doesn't contain any role prefixes, assume it's a user message
		if !strings.Contains(trimmedText, "user:") &&
			!strings.Contains(trimmedText, "assistant:") &&
			!strings.Contains(trimmedText, "system:") {
			return trimmedText
		}
	}

	// If still no user message found, return the original text as fallback
	return conversationText
}

// TraceGeneration traces an LLM generation using OTEL spans
func (t *OTELTracer) TraceGeneration(ctx context.Context, modelName string, prompt string, response string, startTime time.Time, endTime time.Time, metadata map[string]interface{}) (string, error) {
	if !t.enabled {
		return "", nil
	}

	// Get organization ID from context
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Get agent name from context if available
	agentName, _ := telemetry.GetAgentName(ctx)

	// Get span name from agent context or use default
	spanName := telemetry.GetSpanNameOrDefault(ctx, "llm.generation")

	// Check for tool calls from context
	toolCalls := telemetry.GetToolCallsFromContext(ctx)

	var outputWithToolCalls string
	if len(toolCalls) > 0 {
		// Try to parse response as JSON and add tool_calls field
		var responseObj map[string]interface{}
		if err := json.Unmarshal([]byte(response), &responseObj); err == nil {
			// Successfully parsed as JSON, add tool_calls field
			responseObj["tool_calls"] = toolCalls
			if modifiedJSON, err := json.Marshal(responseObj); err == nil {
				outputWithToolCalls = string(modifiedJSON)
			} else {
				// Fallback to original response if marshaling fails
				outputWithToolCalls = response
			}
		} else {
			// Not valid JSON, fallback to text concatenation
			toolCallsJSON, _ := json.MarshalIndent(toolCalls, "", "  ")
			outputWithToolCalls = fmt.Sprintf("%s\n\n**Tool Calls:**\n```json\n%s\n```", response, string(toolCallsJSON))
		}
	} else {
		outputWithToolCalls = response
	}

	// Build attributes for the generation span
	attrs := []attribute.KeyValue{
		// GenAI semantic conventions that Langfuse expects
		attribute.String("gen_ai.request.model", modelName),
		attribute.String("gen_ai.system", "openai"), // Can be made configurable

		// Langfuse-specific trace-level attributes (for list view)
		attribute.String("langfuse.trace.name", telemetry.GetTraceNameOrDefault(ctx, spanName)),
		attribute.String("langfuse.trace.input", prompt),
		attribute.String("langfuse.trace.output", outputWithToolCalls),

		// Langfuse-specific observation-level attributes (for detailed view)
		attribute.String("langfuse.environment", t.config.Environment),
		attribute.String("langfuse.observation.type", "generation"),
		attribute.String("langfuse.observation.input", prompt),
		attribute.String("langfuse.observation.output", outputWithToolCalls),

		// Token usage with proper GenAI attributes (based on last user message only)
		attribute.Int64("gen_ai.usage.prompt_tokens", int64(len(prompt)/4)), // Rough estimate
		attribute.Int64("gen_ai.usage.completion_tokens", int64(len(response)/4)),
		attribute.Int64("gen_ai.usage.total_tokens", int64((len(prompt)+len(response))/4)),
	}

	// Add organization ID if available
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

	// Add prompt attributes using the last user message
	promptAttrs := t.promptToAttributes(prompt)
	attrs = append(attrs, promptAttrs...)

	// Add response attributes
	responseAttrs := t.responseToAttributes(response)
	attrs = append(attrs, responseAttrs...)

	// Create LLM generation span (will be child of existing span if one exists)
	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithTimestamp(startTime),
		trace.WithAttributes(attrs...),
	)
	defer span.End(trace.WithTimestamp(endTime))

	// Create individual spans for each tool call at the trace level (not as children)
	if len(toolCalls) > 0 {
		// Create tool call spans using the root context to make them appear as separate timeline items
		t.createToolCallSpansAsTraceItems(ctx, toolCalls)

		// Also add tool calls to metadata for backward compatibility
		for i, toolCall := range toolCalls {
			prefix := fmt.Sprintf("tool_call_%d", i)
			span.SetAttributes(
				attribute.String("langfuse.observation.metadata."+prefix+".name", toolCall.Name),
				attribute.String("langfuse.observation.metadata."+prefix+".arguments", toolCall.Arguments),
				attribute.String("langfuse.observation.metadata."+prefix+".result", toolCall.Result),
			)
			if toolCall.Error != "" {
				span.SetAttributes(attribute.String("langfuse.observation.metadata."+prefix+".error", toolCall.Error))
			}
		}
		span.SetAttributes(attribute.Int("langfuse.observation.metadata.tool_calls_count", len(toolCalls)))
	}

	// Add metadata as span attributes using proper Langfuse namespace
	for k, v := range metadata {
		span.SetAttributes(attribute.String("langfuse.observation.metadata."+k, fmt.Sprintf("%v", v)))
	}

	return span.SpanContext().SpanID().String(), nil
}

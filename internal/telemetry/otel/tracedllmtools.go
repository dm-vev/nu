package otel

import (
	"context"
	"strings"
	"time"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// GenerateWithTools generates text from a prompt with tools using unified tracing
func (m *TracedLLM) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	// First check if underlying LLM supports GenerateWithTools
	if llmWithTools, ok := m.llm.(interface {
		GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error)
	}); ok {
		startTime := time.Now()

		// Start span
		ctx, span := m.tracer.StartSpan(ctx, "llm.generate_with_tools")
		defer span.End()

		// Add attributes
		span.SetAttribute("prompt.length", len(prompt))
		span.SetAttribute("prompt.hash", hashString(prompt))
		span.SetAttribute("tools.count", len(tools))

		// Extract model name from LLM client
		model := "unknown"
		if modelProvider, ok := m.llm.(interface{ GetModel() string }); ok {
			model = modelProvider.GetModel()
		}
		if model == "" {
			model = m.llm.Name() // fallback to provider name
		}
		span.SetAttribute("model", model)

		// Add tool names if available
		if len(tools) > 0 {
			toolNames := make([]string, len(tools))
			for i, tool := range tools {
				toolNames[i] = tool.Name()
			}
			span.SetAttribute("tools", strings.Join(toolNames, ","))
		}

		// Initialize tool calls collection in context for tracing
		ctx = telemetry.WithToolCallsCollection(ctx)

		// Call the underlying LLM's GenerateWithTools method
		response, err := llmWithTools.GenerateWithTools(ctx, prompt, tools, options...)

		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// Add response attributes
		if err == nil {
			span.SetAttribute("response.length", len(response))
			span.SetAttribute("response.hash", hashString(response))
			span.SetAttribute("duration_ms", duration.Milliseconds())

			// Include actual content if configured
			if m.shouldIncludeContent() {
				span.SetAttribute("prompt.content", prompt)
				span.SetAttribute("response.content", response)
			}
		} else {
			span.RecordError(err)
		}

		// Mirror the streaming path: if any tool calls were collected, emit a
		// generation span via TraceGeneration so Langfuse / OTEL backends record
		// the call graph for non-streaming runs too (#295).
		if toolCalls := telemetry.GetToolCallsFromContext(ctx); len(toolCalls) > 0 {
			responseText := response
			if !m.shouldIncludeContent() {
				responseText = "<redacted>"
			}
			metadata := map[string]any{
				"streaming": false,
				"tools":     len(tools),
			}
			// Stamp the error so downstream backends can distinguish a
			// failed generation from a successful one with empty content.
			if err != nil {
				metadata["error"] = err.Error()
				if responseText == "" {
					responseText = "<error>"
				}
			}
			if tracer, ok := m.tracer.(interface {
				TraceGeneration(context.Context, string, string, string, time.Time, time.Time, map[string]interface{}) (string, error)
			}); ok {
				_, _ = tracer.TraceGeneration(ctx, model, prompt, responseText, startTime, endTime, metadata) //nolint:gosec
			}
		}

		return response, err
	}

	// Fallback to regular Generate if GenerateWithTools is not supported
	return m.Generate(ctx, prompt, options...)
}

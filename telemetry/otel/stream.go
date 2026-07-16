package otel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

// GenerateStream implements contracts.StreamingLLM.GenerateStream
func (m *TracedLLM) GenerateStream(ctx context.Context, prompt string, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	// Check if underlying LLM supports streaming
	streamingLLM, ok := m.llm.(contracts.StreamingLLM)
	if !ok {
		return nil, fmt.Errorf("underlying LLM does not support streaming")
	}

	// Start span
	ctx, span := m.tracer.StartSpan(ctx, "llm.generate_stream")
	defer span.End()

	// Add attributes
	span.SetAttribute("prompt.length", len(prompt))
	span.SetAttribute("prompt.hash", hashString(prompt))
	span.SetAttribute("streaming", true)

	// Extract model name from LLM client
	model := "unknown"
	if modelProvider, ok := m.llm.(interface{ GetModel() string }); ok {
		model = modelProvider.GetModel()
	}
	if model == "" {
		model = m.llm.Name() // fallback to provider name
	}
	span.SetAttribute("model", model)

	// Include actual prompt content if configured (response is streamed)
	if m.shouldIncludeContent() {
		span.SetAttribute("prompt.content", prompt)
	}

	return streamingLLM.GenerateStream(ctx, prompt, options...)
}

// GenerateWithToolsStream implements contracts.StreamingLLM.GenerateWithToolsStream
func (m *TracedLLM) GenerateWithToolsStream(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	// Check if underlying LLM supports streaming with tools
	streamingLLM, ok := m.llm.(contracts.StreamingLLM)
	if !ok {
		return nil, fmt.Errorf("underlying LLM does not support streaming")
	}

	// Start span
	startTime := time.Now()
	ctx, span := m.tracer.StartSpan(ctx, "llm.generate_with_tools_stream")

	// Add attributes
	span.SetAttribute("prompt.length", len(prompt))
	span.SetAttribute("prompt.hash", hashString(prompt))
	span.SetAttribute("streaming", true)
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

	// Include actual prompt content if configured (response is streamed)
	if m.shouldIncludeContent() {
		span.SetAttribute("prompt.content", prompt)
	}

	// Initialize tool calls collection in context for tracing
	ctx = telemetry.WithToolCallsCollection(ctx)

	// Get the original stream
	originalChan, err := streamingLLM.GenerateWithToolsStream(ctx, prompt, tools, options...)
	if err != nil {
		span.RecordError(err)
		span.End()
		return nil, err
	}

	// Create a new channel to wrap the original
	wrappedChan := make(chan contracts.StreamEvent, 10)

	// Start a goroutine to proxy events and handle span completion
	go func() {
		defer close(wrappedChan)
		defer func() {
			// When streaming is complete, create tool call spans and end main span
			endTime := time.Now()
			duration := endTime.Sub(startTime)
			span.SetAttribute("duration_ms", duration.Milliseconds())

			// Get tool calls from context and create spans using TraceGeneration if any exist
			toolCalls := telemetry.GetToolCallsFromContext(ctx)

			if len(toolCalls) > 0 {
				// Extract model name
				model := "unknown"
				if modelProvider, ok := streamingLLM.(interface{ GetModel() string }); ok {
					model = modelProvider.GetModel()
				}
				if model == "" {
					model = streamingLLM.Name()
				}

				// Create spans using TraceGeneration which handles tool calls correctly
				if tracer, ok := m.tracer.(interface {
					TraceGeneration(context.Context, string, string, string, time.Time, time.Time, map[string]interface{}) (string, error)
				}); ok {
					_, _ = tracer.TraceGeneration(ctx, model, prompt, "streaming_response", startTime, endTime, map[string]any{ //nolint:gosec
						"streaming": true,
						"tools":     len(tools),
					})
				}
			}

			span.End()
		}()

		// Proxy all events from the original channel
		for event := range originalChan {
			wrappedChan <- event
		}
	}()

	return wrappedChan, nil
}

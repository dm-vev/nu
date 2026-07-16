package otel

import (
	"context"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"
)

// GenerateDetailed generates text and returns detailed response information including token usage
func (m *TracedLLM) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	startTime := time.Now()

	// Start span
	ctx, span := m.tracer.StartSpan(ctx, "llm.generate_detailed")
	defer span.End()

	// Add attributes
	span.SetAttribute("prompt.length", len(prompt))
	span.SetAttribute("prompt.hash", hashString(prompt))

	// Extract model name from LLM client
	model := "unknown"
	if modelProvider, ok := m.llm.(interface{ GetModel() string }); ok {
		model = modelProvider.GetModel()
	}
	if model == "" {
		model = m.llm.Name() // fallback to provider name
	}
	span.SetAttribute("model", model)

	// Call the underlying LLM
	response, err := m.llm.GenerateDetailed(ctx, prompt, options...)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Set response attributes
	span.SetAttribute("response.length", len(response.Content))
	span.SetAttribute("response.model", response.Model)
	if response.Usage != nil {
		span.SetAttribute("usage.input_tokens", response.Usage.InputTokens)
		span.SetAttribute("usage.output_tokens", response.Usage.OutputTokens)
		span.SetAttribute("usage.total_tokens", response.Usage.TotalTokens)
		if response.Usage.ReasoningTokens > 0 {
			span.SetAttribute("usage.reasoning_tokens", response.Usage.ReasoningTokens)
		}
	}
	span.SetAttribute("duration_ms", duration.Milliseconds())

	// Include actual content if configured
	if m.shouldIncludeContent() {
		span.SetAttribute("prompt.content", prompt)
		span.SetAttribute("response.content", response.Content)
	}

	return response, nil
}

// GenerateWithToolsDetailed generates text with tools and returns detailed response information including token usage
func (m *TracedLLM) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	startTime := time.Now()

	// Start span
	ctx, span := m.tracer.StartSpan(ctx, "llm.generate_with_tools_detailed")
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

	// Add tool names as attributes
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Name()
	}
	span.SetAttribute("tools.names", strings.Join(toolNames, ","))

	// Call the underlying LLM
	response, err := m.llm.GenerateWithToolsDetailed(ctx, prompt, tools, options...)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Set response attributes
	span.SetAttribute("response.length", len(response.Content))
	span.SetAttribute("response.model", response.Model)
	if response.Usage != nil {
		span.SetAttribute("usage.input_tokens", response.Usage.InputTokens)
		span.SetAttribute("usage.output_tokens", response.Usage.OutputTokens)
		span.SetAttribute("usage.total_tokens", response.Usage.TotalTokens)
		if response.Usage.ReasoningTokens > 0 {
			span.SetAttribute("usage.reasoning_tokens", response.Usage.ReasoningTokens)
		}
	}
	span.SetAttribute("duration_ms", duration.Milliseconds())

	return response, nil
}

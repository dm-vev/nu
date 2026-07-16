package otel

import (
	"context"
	"time"

	"github.com/dm-vev/nu/contracts"
)

// TracedLLM implements middleware for LLM calls with unified tracing
type TracedLLM struct {
	llm    contracts.LLM
	tracer contracts.Tracer
}

// NewTracedLLM creates a new LLM middleware with unified tracing
func NewTracedLLM(llm contracts.LLM, tracer contracts.Tracer) contracts.LLM {
	return &TracedLLM{
		llm:    llm,
		tracer: tracer,
	}
}

// shouldIncludeContent checks if the tracer supports and has enabled content inclusion
func (m *TracedLLM) shouldIncludeContent() bool {
	if tracer, ok := m.tracer.(interface{ ShouldIncludeContent() bool }); ok {
		return tracer.ShouldIncludeContent()
	}
	return false
}

// Generate generates text from a prompt with tracing
func (m *TracedLLM) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	startTime := time.Now()

	// Start span
	ctx, span := m.tracer.StartSpan(ctx, "llm.generate")
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
	response, err := m.llm.Generate(ctx, prompt, options...)

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

	return response, err
}

// Name implements contracts.LLM.Name
func (m *TracedLLM) Name() string {
	return m.llm.Name()
}

// SupportsStreaming implements contracts.LLM.SupportsStreaming
func (m *TracedLLM) SupportsStreaming() bool {
	return m.llm.SupportsStreaming()
}

// GetModel returns the model name from the underlying LLM
func (m *TracedLLM) GetModel() string {
	if modelProvider, ok := m.llm.(interface{ GetModel() string }); ok {
		return modelProvider.GetModel()
	}
	// Fallback to provider name if GetModel is not available
	return m.llm.Name()
}

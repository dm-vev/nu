package guardrails

import (
	"context"

	"github.com/dm-vev/nu/contracts"
)

// GuardrailToolMiddleware applies guardrails around tool calls.
type GuardrailToolMiddleware struct {
	tool     contracts.Tool
	pipeline *GuardrailPipeline
}

// NewGuardrailToolMiddleware creates guarded tool middleware.
func NewGuardrailToolMiddleware(tool contracts.Tool, pipeline *GuardrailPipeline) *GuardrailToolMiddleware {
	return &GuardrailToolMiddleware{
		tool:     tool,
		pipeline: pipeline,
	}
}

// Name returns the name of the tool
func (m *GuardrailToolMiddleware) Name() string {
	return m.tool.Name()
}

// Description returns a description of what the tool does
func (m *GuardrailToolMiddleware) Description() string {
	return m.tool.Description()
}

// Parameters returns the parameters that the tool accepts
func (m *GuardrailToolMiddleware) Parameters() map[string]contracts.ParameterSpec {
	return m.tool.Parameters()
}

// Run executes the tool with the given input
func (m *GuardrailToolMiddleware) Run(ctx context.Context, input string) (string, error) {
	// Process request through guardrails
	processedInput, err := m.pipeline.ProcessRequest(ctx, input)
	if err != nil {
		return "", err
	}

	// Call the underlying tool
	output, err := m.tool.Run(ctx, processedInput)
	if err != nil {
		return "", err
	}

	// Process response through guardrails
	processedOutput, err := m.pipeline.ProcessResponse(ctx, output)
	if err != nil {
		return "", err
	}

	return processedOutput, nil
}

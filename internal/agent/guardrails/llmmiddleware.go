package guardrails

import (
	"context"

	"nu/internal/contracts"
)

// GuardrailLLMMiddleware applies guardrails around LLM calls.
type GuardrailLLMMiddleware struct {
	llm      contracts.LLM
	pipeline *GuardrailPipeline
}

// NewGuardrailLLMMiddleware creates guarded LLM middleware.
func NewGuardrailLLMMiddleware(llm contracts.LLM, pipeline *GuardrailPipeline) *GuardrailLLMMiddleware {
	return &GuardrailLLMMiddleware{
		llm:      llm,
		pipeline: pipeline,
	}
}

// Generate generates text from a prompt
func (m *GuardrailLLMMiddleware) Generate(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	// Process request through guardrails
	processedPrompt, err := m.pipeline.ProcessRequest(ctx, prompt)
	if err != nil {
		return "", err
	}

	// Call the underlying LLM
	// Pass an empty slice of options instead of nil to avoid nil pointer dereference
	response, err := m.llm.Generate(ctx, processedPrompt, []contracts.GenerateOption{}...)
	if err != nil {
		return "", err
	}

	// Process response through guardrails
	processedResponse, err := m.pipeline.ProcessResponse(ctx, response)
	if err != nil {
		return "", err
	}

	return processedResponse, nil
}

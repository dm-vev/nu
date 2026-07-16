package contracts

import "context"

// Guardrails represents a system for ensuring safe and appropriate responses
type Guardrails interface {
	// ProcessInput processes user input before sending to the LLM
	ProcessInput(ctx context.Context, input string) (string, error)

	// ProcessOutput processes LLM output before returning to the user
	ProcessOutput(ctx context.Context, output string) (string, error)
}

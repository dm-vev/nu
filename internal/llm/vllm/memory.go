package vllm

import (
	"context"

	"nu/internal/contracts"
	"nu/internal/memory"
)

// buildPromptWithMemory builds a prompt with memory context for prompt-based models
func (c *Client) buildPromptWithMemory(ctx context.Context, prompt string, params *contracts.GenerateOptions) string {
	return memory.BuildInlineHistoryPrompt(ctx, prompt, params.Memory, c.logger)
}

package vllm

import (
	"context"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/history"
)

// buildPromptWithMemory builds a prompt with memory context for prompt-based models
func (c *Client) buildPromptWithMemory(ctx context.Context, prompt string, params *contracts.GenerateOptions) string {
	return history.BuildInlineHistoryPrompt(ctx, prompt, params.Memory, c.logger)
}

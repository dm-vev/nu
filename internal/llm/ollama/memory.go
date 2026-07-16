package ollama

import (
	"context"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/history"
)

// buildPromptWithMemory builds a prompt with memory context for prompt-based models
func (c *Client) buildPromptWithMemory(ctx context.Context, prompt string, params *contracts.GenerateOptions) string {
	return history.BuildInlineHistoryPrompt(ctx, prompt, params.Memory, c.logger)
}

// ollamaPersistToolResultMessage records a tool result message in Memory so the
// next agent turn can replay the tool exchange. callID is synthesized by
// the caller per invocation (not per tool name) so the same tool called
// twice in one assistant turn doesn't share an ID. BuildInlineHistoryPrompt
// requires a non-empty ToolCallID to render the message back into the
// inlined history on the next turn.
func ollamaPersistToolResultMessage(ctx context.Context, mem contracts.Memory, callID, toolName, content string) {
	if mem == nil {
		return
	}
	_ = mem.AddMessage(ctx, contracts.Message{
		Role:       contracts.MessageRoleTool,
		Content:    content,
		ToolCallID: callID,
		Metadata: map[string]interface{}{
			"tool_name": toolName,
		},
	})
}

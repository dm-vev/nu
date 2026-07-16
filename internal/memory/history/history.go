package history

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

// BuildInlineHistoryPrompt builds a prompt with conversation history for prompt-based models
// This is a generic implementation for models that don't have native conversation support
func BuildInlineHistoryPrompt(ctx context.Context, prompt string, memory contracts.Memory, logger telemetry.Logger) string {
	if memory == nil {
		return "User: " + prompt
	}

	memoryMessages, err := memory.GetMessages(ctx)
	if err != nil {
		if logger != nil {
			logger.Error(ctx, "Failed to retrieve memory messages", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return "User: " + prompt
	}

	if len(memoryMessages) == 0 {
		return prompt
	}

	// Format memory messages into prompt for prompt-based models, ensuring system messages come first
	var promptBuilder strings.Builder

	// Add system messages first
	for _, msg := range memoryMessages {
		if msg.Role == contracts.MessageRoleSystem {
			promptBuilder.WriteString("System: " + msg.Content + "\n")
		}
	}

	// Add other messages in order
	for _, msg := range memoryMessages {
		switch msg.Role {
		case contracts.RoleUser:
			promptBuilder.WriteString("User: " + msg.Content + "\n")
		case contracts.RoleAssistant:
			if msg.Content != "" {
				promptBuilder.WriteString("Assistant: " + msg.Content + "\n")
			}
		case contracts.MessageRoleTool:
			if msg.ToolCallID != "" {
				toolName := "unknown"
				if msg.Metadata != nil {
					if name, ok := msg.Metadata["tool_name"].(string); ok {
						toolName = name
					}
				}
				fmt.Fprintf(&promptBuilder, "Tool %s result: %s\n", toolName, msg.Content)
			}
		}
	}
	promptBuilder.WriteString("User: " + prompt)
	return promptBuilder.String()
}

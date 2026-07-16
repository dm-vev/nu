package generation

import (
	"context"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

func saveStreamMessages(ctx context.Context, memory contracts.Memory, logger telemetry.Logger, content string, toolCalls []contracts.ToolCall, toolResults map[string]string) {
	if memory == nil {
		return
	}
	if len(toolCalls) == 0 {
		if content != "" {
			addStreamMessage(ctx, memory, logger, contracts.Message{Role: "assistant", Content: content}, "assistant response")
		}
		return
	}

	addStreamMessage(ctx, memory, logger, contracts.Message{Role: "assistant", Content: content, ToolCalls: toolCalls}, "assistant tool calls")
	for _, toolCall := range toolCalls {
		result, ok := toolResults[toolCall.ID]
		if !ok {
			continue
		}
		addStreamMessage(ctx, memory, logger, contracts.Message{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCall.ID,
			Metadata:   map[string]interface{}{"tool_name": toolCall.Name},
		}, "tool result")
	}
}

func addStreamMessage(ctx context.Context, memory contracts.Memory, logger telemetry.Logger, message contracts.Message, label string) {
	if err := memory.AddMessage(ctx, message); err != nil && logger != nil {
		logger.Warn(ctx, "Failed to add "+label+" to memory", map[string]interface{}{"error": err.Error()})
	}
}

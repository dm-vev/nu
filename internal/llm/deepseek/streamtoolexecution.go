package deepseek

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
)

func (c *Client) executeStreamTools(
	ctx context.Context,
	tools []contracts.Tool,
	toolCalls []ToolCall,
	iteration int,
	eventChan chan contracts.StreamEvent,
) []Message {
	var messages []Message

	// Process each tool call
	for _, toolCall := range toolCalls {
		// Find the matching tool
		var foundTool contracts.Tool
		for _, tool := range tools {
			if tool.Name() == toolCall.Function.Name {
				foundTool = tool
				break
			}
		}

		if foundTool == nil {
			c.logger.Error(ctx, "Tool not found", map[string]interface{}{
				"tool_name": toolCall.Function.Name,
			})
			continue
		}

		// Execute the tool
		result, err := foundTool.Execute(ctx, toolCall.Function.Arguments)
		if err != nil {
			c.logger.Error(ctx, "Tool execution error", map[string]interface{}{
				"tool_name": toolCall.Function.Name,
				"error":     err.Error(),
			})
			result = fmt.Sprintf("Error executing tool: %v", err)
		}

		// Send tool result event
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventToolResult,
			Timestamp: time.Now(),
			ToolCall: &contracts.ToolCall{
				ID:        toolCall.ID,
				Name:      toolCall.Function.Name,
				Arguments: toolCall.Function.Arguments,
			},
			Metadata: map[string]interface{}{
				"iteration": iteration + 1,
				"result":    result,
			},
		}

		// Add the tool result to the conversation
		c.logger.Debug(ctx, "Adding tool result to conversation", map[string]interface{}{
			"tool_call_id":  toolCall.ID,
			"tool_name":     toolCall.Function.Name,
			"result_length": len(result),
		})

		messages = append(messages, Message{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCall.ID,
		})
	}

	return messages
}

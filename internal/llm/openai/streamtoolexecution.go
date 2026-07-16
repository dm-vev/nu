package openai

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"

	"github.com/openai/openai-go/v2"
)

func (c *Client) executeStreamToolCalls(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	toolCalls []openai.ChatCompletionMessageToolCallUnion,
	tools []contracts.Tool,
	iteration int,
	eventChan chan<- contracts.StreamEvent,
) []openai.ChatCompletionMessageParamUnion {
	for _, toolCall := range toolCalls {
		var foundTool contracts.Tool
		for _, tool := range tools {
			if tool.Name() == toolCall.Function.Name {
				foundTool = tool
				break
			}
		}
		if foundTool == nil {
			c.logger.Error(ctx, "Tool not found", map[string]interface{}{"tool_name": toolCall.Function.Name})
			continue
		}
		result, err := foundTool.Execute(ctx, toolCall.Function.Arguments)
		if err != nil {
			c.logger.Error(ctx, "Tool execution error", map[string]interface{}{"tool_name": toolCall.Function.Name, "error": err.Error()})
			result = fmt.Sprintf("Error executing tool: %v", err)
		}
		eventChan <- contracts.StreamEvent{
			Type: contracts.StreamEventToolResult, Timestamp: time.Now(), Content: result,
			ToolCall: &contracts.ToolCall{ID: toolCall.ID, Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments},
			Metadata: map[string]interface{}{"iteration": iteration + 1, "result": result},
		}
		c.logger.Debug(ctx, "Adding tool result to conversation", map[string]interface{}{
			"tool_call_id": toolCall.ID, "id_length": len(toolCall.ID), "tool_name": toolCall.Function.Name, "result_length": len(result),
		})
		toolMessage := openai.ToolMessage(result, toolCall.ID)
		c.logger.Debug(ctx, "Created tool message", map[string]interface{}{"message_type": "tool"})
		messages = append(messages, toolMessage)
	}
	return messages
}

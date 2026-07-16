package anthropic

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
)

func (c *Client) executeStreamingTools(ctx context.Context, toolCalls []contracts.ToolCall, originalTools []contracts.Tool, messages *[]Message, eventChan chan<- contracts.StreamEvent, iteration int) error {
	select {
	case eventChan <- contracts.StreamEvent{
		Type: contracts.StreamEventContentDelta, Content: "\n", Timestamp: time.Now(),
		Metadata: map[string]interface{}{"before_tools": true, "iteration": iteration + 1},
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	for _, toolCall := range toolCalls {
		var selectedTool contracts.Tool
		for _, tool := range originalTools {
			if tool.Name() == toolCall.Name {
				selectedTool = tool
				break
			}
		}
		if selectedTool == nil {
			c.logger.Error(ctx, "Tool not found in streaming", map[string]interface{}{"toolName": toolCall.Name})
			errorMessage := fmt.Sprintf("Error: tool not found: %s", toolCall.Name)
			*messages = append(*messages, Message{Role: "user", Content: fmt.Sprintf("Tool %s result: %s", toolCall.Name, errorMessage)})
			select {
			case eventChan <- contracts.StreamEvent{
				Type:     contracts.StreamEventToolResult,
				ToolCall: &contracts.ToolCall{ID: toolCall.ID, Name: toolCall.Name, Arguments: toolCall.Arguments},
				Content:  errorMessage, Timestamp: time.Now(),
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}

		c.logger.Info(ctx, "[TOOL EXECUTION DEBUG] Executing tool in streaming", map[string]interface{}{
			"toolName": toolCall.Name, "arguments": toolCall.Arguments, "iteration": iteration + 1,
		})
		toolResult, err := selectedTool.Execute(ctx, toolCall.Arguments)
		if err != nil {
			toolResult = fmt.Sprintf("Error: %v", err)
		}
		*messages = append(*messages, Message{Role: "user", Content: fmt.Sprintf("Tool %s result: %s", toolCall.Name, toolResult)})
		select {
		case eventChan <- contracts.StreamEvent{
			Type:     contracts.StreamEventToolResult,
			ToolCall: &contracts.ToolCall{ID: toolCall.ID, Name: toolCall.Name, Arguments: toolCall.Arguments},
			Content:  toolResult, Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

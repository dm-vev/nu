package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/genai"

	"github.com/dm-vev/nu/contracts"
)

// streamResponse streams a response string in chunks
func (c *Client) streamResponse(ctx context.Context, response string, eventCh chan contracts.StreamEvent) {
	chunkSize := 50 // characters per chunk
	for i := 0; i < len(response); i += chunkSize {
		end := i + chunkSize
		if end > len(response) {
			end = len(response)
		}

		chunk := response[i:end]

		// Send content delta event
		select {
		case eventCh <- contracts.StreamEvent{
			Type:      contracts.StreamEventContentDelta,
			Content:   chunk,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}

		// No artificial delay - stream at real speed
	}
}

func (c *Client) executeStreamToolCalls(
	ctx context.Context,
	contents []*genai.Content,
	toolCalls []contracts.ToolCall,
	tools []contracts.Tool,
	iteration int,
	eventCh chan contracts.StreamEvent,
) ([]*genai.Content, error) {
	c.logger.Info(ctx, "Processing tool calls in streaming", map[string]interface{}{
		"count":     len(toolCalls),
		"iteration": iteration + 1,
	})

	// Add assistant message with tool calls
	assistantMessage := &genai.Content{
		Role:  "model",
		Parts: []*genai.Part{},
	}

	// Convert tool calls to Gemini format and add to message
	for _, toolCall := range toolCalls {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Arguments), &args); err != nil {
			args = make(map[string]interface{})
		}
		assistantMessage.Parts = append(assistantMessage.Parts, &genai.Part{
			ThoughtSignature: toolCall.ThoughtSignature,
			FunctionCall: &genai.FunctionCall{
				Name: toolCall.Name,
				Args: args,
			},
		})
	}

	contents = append(contents, assistantMessage)

	// Collect all tool results to add them in a single content message
	var functionResponses []*genai.Part

	// Execute each tool and collect results
	for _, toolCall := range toolCalls {
		// Find the requested tool
		var selectedTool contracts.Tool
		for _, tool := range tools {
			if tool.Name() == toolCall.Name {
				selectedTool = tool
				break
			}
		}

		if selectedTool == nil {
			c.logger.Error(ctx, "Tool not found in streaming", map[string]interface{}{
				"toolName": toolCall.Name,
			})

			// Add tool not found error as function response
			errorMessage := fmt.Sprintf("Error: tool not found: %s", toolCall.Name)
			functionResponses = append(functionResponses, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name: toolCall.Name,
					Response: map[string]any{
						"error": errorMessage,
					},
				},
			})

			// Send tool result event with error
			select {
			case eventCh <- contracts.StreamEvent{
				Type: contracts.StreamEventToolResult,
				ToolCall: &contracts.ToolCall{
					ID:        toolCall.ID,
					Name:      toolCall.Name,
					Arguments: toolCall.Arguments,
				},
				Content:   errorMessage,
				Timestamp: time.Now(),
			}:
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			continue
		}

		// Execute tool
		c.logger.Info(ctx, "Executing tool in streaming", map[string]interface{}{
			"toolName":  toolCall.Name,
			"arguments": toolCall.Arguments,
			"iteration": iteration + 1,
		})

		toolResult, err := selectedTool.Execute(ctx, toolCall.Arguments)
		if err != nil {
			toolResult = fmt.Sprintf("Error: %v", err)
		}

		// Add tool result as function response
		functionResponses = append(functionResponses, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				Name: toolCall.Name,
				Response: map[string]any{
					"result": toolResult,
				},
			},
		})

		// Send tool result event
		select {
		case eventCh <- contracts.StreamEvent{
			Type: contracts.StreamEventToolResult,
			ToolCall: &contracts.ToolCall{
				ID:        toolCall.ID,
				Name:      toolCall.Name,
				Arguments: toolCall.Arguments,
			},
			Content:   toolResult, // Tool result goes in Content field
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Add all function responses in a single content message
	if len(functionResponses) > 0 {
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: functionResponses,
		})
	}

	return contents, nil
}

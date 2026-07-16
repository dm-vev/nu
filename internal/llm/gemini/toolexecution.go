package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"google.golang.org/genai"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

func (c *Client) executeToolCalls(
	ctx context.Context,
	contents []*genai.Content,
	parts []*genai.Part,
	tools []contracts.Tool,
	params *contracts.GenerateOptions,
	toolCallHistory map[string]int,
	toolCallHistoryMu *sync.Mutex,
	iteration int,
) ([]*genai.Content, error) {
	c.logger.Info(ctx, "Processing function calls", map[string]interface{}{
		"iteration": iteration + 1,
	})

	// Add the assistant's message with function calls to the conversation
	// Ensure the role is set to "model"
	contents = append(contents, &genai.Content{
		Role:  "model",
		Parts: parts,
	})

	// Collect all function responses to add them in a single content message
	var functionResponses []*genai.Part

	// Process each function call
	for _, part := range parts {
		if part.FunctionCall == nil {
			continue
		}

		functionCall := part.FunctionCall

		// Find the requested tool
		var selectedTool contracts.Tool
		for _, tool := range tools {
			if tool.Name() == functionCall.Name {
				selectedTool = tool
				break
			}
		}

		if selectedTool == nil {
			c.logger.Error(ctx, "Tool not found", map[string]interface{}{
				"toolName": functionCall.Name,
			})

			// Add tool not found error as function response
			functionResponses = append(functionResponses, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name: functionCall.Name,
					Response: map[string]any{
						"error": fmt.Sprintf("tool not found: %s", functionCall.Name),
					},
				},
			})

			// Store failed tool call in memory if provided
			if params.Memory != nil {
				_ = params.Memory.AddMessage(ctx, contracts.Message{
					Role:    "assistant",
					Content: "",
					ToolCalls: []contracts.ToolCall{{
						Name:             functionCall.Name,
						Arguments:        "{}",
						ThoughtSignature: part.ThoughtSignature,
					}},
				})
				_ = params.Memory.AddMessage(ctx, contracts.Message{
					Role:    "tool",
					Content: fmt.Sprintf("Error: tool not found: %s", functionCall.Name),
					Metadata: map[string]interface{}{
						"tool_name": functionCall.Name,
					},
				})
			}

			// Add to tracing context
			toolCallTrace := telemetry.ToolCall{
				Name:       functionCall.Name,
				Arguments:  "{}",
				Timestamp:  time.Now().Format(time.RFC3339),
				StartTime:  time.Now(),
				Duration:   0,
				DurationMs: 0,
				Error:      fmt.Sprintf("tool not found: %s", functionCall.Name),
				Result:     fmt.Sprintf("Error: tool not found: %s", functionCall.Name),
			}

			telemetry.AddToolCallToContext(ctx, toolCallTrace)
			continue
		}

		// Convert function call arguments to JSON string
		argsBytes, err := json.Marshal(functionCall.Args)
		if err != nil {
			c.logger.Error(ctx, "Failed to marshal function call arguments", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to marshal function call arguments: %w", err)
		}

		// Execute the tool
		c.logger.Info(ctx, "Executing tool", map[string]interface{}{"toolName": selectedTool.Name()})
		toolStartTime := time.Now()
		toolResult, err := selectedTool.Execute(ctx, string(argsBytes))
		toolEndTime := time.Now()

		// Check for repetitive calls and add warning if needed
		cacheKey := functionCall.Name + ":" + string(argsBytes)
		toolCallHistoryMu.Lock()
		toolCallHistory[cacheKey]++
		callCount := toolCallHistory[cacheKey]
		toolCallHistoryMu.Unlock()

		if callCount > 1 {
			warning := fmt.Sprintf("\n\n[WARNING: This is call #%d to %s with identical parameters. You may be in a loop. Consider using the available information to provide a final answer.]",
				callCount,
				functionCall.Name)
			if err == nil {
				toolResult += warning
			}
			c.logger.Warn(ctx, "Repetitive tool call detected", map[string]interface{}{
				"toolName":  functionCall.Name,
				"callCount": callCount,
			})
		}

		// Add tool call to tracing context
		executionDuration := toolEndTime.Sub(toolStartTime)
		toolCallTrace := telemetry.ToolCall{
			Name:       functionCall.Name,
			Arguments:  string(argsBytes),
			Timestamp:  toolStartTime.Format(time.RFC3339),
			StartTime:  toolStartTime,
			Duration:   executionDuration,
			DurationMs: executionDuration.Milliseconds(),
		}

		// Store tool call and result in memory if provided
		if params.Memory != nil {
			_ = params.Memory.AddMessage(ctx, contracts.Message{
				Role:    "assistant",
				Content: "",
				ToolCalls: []contracts.ToolCall{{
					Name:             functionCall.Name,
					Arguments:        string(argsBytes),
					ThoughtSignature: part.ThoughtSignature,
				}},
			})
			toolMessage := contracts.Message{
				Role: "tool",
				Metadata: map[string]interface{}{
					"tool_name": functionCall.Name,
				},
			}
			if err != nil {
				toolMessage.Content = fmt.Sprintf("Error: %v", err)
			} else {
				toolMessage.Content = toolResult
			}
			_ = params.Memory.AddMessage(ctx, toolMessage)
		}

		if err != nil {
			c.logger.Error(ctx, "Tool execution failed", map[string]interface{}{
				"toolName": selectedTool.Name(),
				"toolArgs": string(argsBytes),
				"error":    err.Error(),
				"duration": toolEndTime.Sub(toolStartTime).String(),
			})
			toolCallTrace.Error = err.Error()
			toolCallTrace.Result = fmt.Sprintf("Error: %v", err)
			functionResponses = append(functionResponses, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name: functionCall.Name,
					Response: map[string]any{
						"error": err.Error(),
					},
				},
			})
		} else {
			toolCallTrace.Result = toolResult
			functionResponses = append(functionResponses, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name: functionCall.Name,
					Response: map[string]any{
						"result": toolResult,
					},
				},
			})
		}

		telemetry.AddToolCallToContext(ctx, toolCallTrace)
	}

	if len(functionResponses) > 0 {
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: functionResponses,
		})
	}

	return contents, nil
}

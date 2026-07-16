package deepseek

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nu/internal/contracts"
)

func (c *Client) streamToolsIteration(
	ctx context.Context,
	params *contracts.GenerateOptions,
	deepseekTools []Tool,
	messages []Message,
	iteration int,
	maxIterations int,
	filterIntermediateContent bool,
	eventChan chan contracts.StreamEvent,
) (deepSeekStreamToolIterationResult, bool) {
	result := deepSeekStreamToolIterationResult{
		assistantResponse: Message{Role: "assistant"},
	}

	// Create request for this iteration
	req := ChatCompletionRequest{
		Model:      c.Model,
		Messages:   messages,
		Tools:      deepseekTools,
		ToolChoice: "auto",
		Stream:     true,
	}

	if params.LLMConfig != nil {
		req.Temperature = params.LLMConfig.Temperature
		req.TopP = params.LLMConfig.TopP
		req.FrequencyPenalty = params.LLMConfig.FrequencyPenalty
		req.PresencePenalty = params.LLMConfig.PresencePenalty
	}

	c.logger.Debug(ctx, "Creating DeepSeek streaming request with tools", map[string]interface{}{
		"model":         c.Model,
		"tools":         len(deepseekTools),
		"temperature":   params.LLMConfig.Temperature,
		"iteration":     iteration + 1,
		"maxIterations": maxIterations,
		"message_count": len(messages),
	})

	// Make streaming HTTP request
	resp, err := c.doStreamRequest(ctx, req)
	if err != nil {
		c.logger.Error(ctx, "Failed to create DeepSeek streaming", map[string]interface{}{
			"error": err.Error(),
		})
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventError,
			Error:     fmt.Errorf("deepseek streaming error: %w", err),
			Timestamp: time.Now(),
		}
		return result, false
	}

	// Track streaming state
	var currentToolCall *contracts.ToolCall
	var toolCallBuffer strings.Builder

	// Process stream chunks
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse SSE format: "data: {json}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end marker
		if data == "[DONE]" {
			break
		}

		// Parse JSON chunk
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			c.logger.Error(ctx, "Failed to parse stream chunk", map[string]interface{}{
				"error": err.Error(),
				"data":  data,
			})
			continue
		}

		// Process choices
		for _, choice := range chunk.Choices {
			// Handle content
			if choice.Delta.Content != "" {
				result.hasContent = true
				result.assistantResponse.Content += choice.Delta.Content

				contentEvent := contracts.StreamEvent{
					Type:      contracts.StreamEventContentDelta,
					Content:   choice.Delta.Content,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"choice_index": choice.Index,
						"iteration":    iteration + 1,
					},
				}

				if filterIntermediateContent && len(deepseekTools) > 0 && iteration < maxIterations-1 {
					// Capture content for potential replay later
					result.contentEvents = append(result.contentEvents, contentEvent)
				} else {
					// Stream content immediately
					eventChan <- contentEvent
				}
			}

			// Handle tool calls - DeepSeek streams them incrementally
			if len(choice.Delta.ToolCalls) > 0 {
				for _, toolCall := range choice.Delta.ToolCalls {
					if toolCall.Function.Name != "" || toolCall.Function.Arguments != "" {
						// Check if this is a new tool call or continuation
						if toolCall.Function.Name != "" {
							// New tool call started
							if currentToolCall != nil && toolCallBuffer.Len() > 0 {
								// Finish previous tool call
								currentToolCall.Arguments = toolCallBuffer.String()
								eventChan <- contracts.StreamEvent{
									Type:      contracts.StreamEventToolUse,
									ToolCall:  currentToolCall,
									Timestamp: time.Now(),
								}
							}

							// Start new tool call
							currentToolCall = &contracts.ToolCall{
								ID:   toolCall.ID,
								Name: toolCall.Function.Name,
							}
							toolCallBuffer.Reset()

							// Add to assistant response
							result.assistantResponse.ToolCalls = append(result.assistantResponse.ToolCalls, ToolCall{
								ID:   toolCall.ID,
								Type: "function",
								Function: FunctionCall{
									Name: toolCall.Function.Name,
								},
							})

							c.logger.Debug(ctx, "Started new tool call", map[string]interface{}{
								"tool_id":   toolCall.ID,
								"tool_name": toolCall.Function.Name,
							})
						}

						// Accumulate arguments
						if toolCall.Function.Arguments != "" {
							toolCallBuffer.WriteString(toolCall.Function.Arguments)
							// Update the last tool call arguments
							if len(result.assistantResponse.ToolCalls) > 0 {
								lastIdx := len(result.assistantResponse.ToolCalls) - 1
								result.assistantResponse.ToolCalls[lastIdx].Function.Arguments += toolCall.Function.Arguments
							}
						}
					}
				}
			}

			// Check for finish reason
			if choice.FinishReason == "tool_calls" && currentToolCall != nil {
				// Finish last tool call
				currentToolCall.Arguments = toolCallBuffer.String()
				eventChan <- contracts.StreamEvent{
					Type:      contracts.StreamEventToolUse,
					ToolCall:  currentToolCall,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"finish_reason": "tool_calls",
						"iteration":     iteration + 1,
					},
				}
				currentToolCall = nil
				toolCallBuffer.Reset()

				c.logger.Debug(ctx, "Finished tool calls", map[string]interface{}{
					"finish_reason": choice.FinishReason,
					"iteration":     iteration + 1,
				})
			}
		}
	}

	if err := resp.Body.Close(); err != nil {
		c.logger.Error(ctx, "Failed to close response body in iteration", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Check for scanner error
	if err := scanner.Err(); err != nil {
		c.logger.Error(ctx, "DeepSeek streaming with tools error", map[string]interface{}{
			"error": err.Error(),
			"model": c.Model,
		})
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventError,
			Error:     fmt.Errorf("deepseek streaming error: %w", err),
			Timestamp: time.Now(),
		}
		return result, false
	}

	return result, true
}

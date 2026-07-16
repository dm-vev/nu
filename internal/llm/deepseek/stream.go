package deepseek

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// GenerateStream implements contracts.StreamingLLM.GenerateStream
func (c *Client) GenerateStream(
	ctx context.Context,
	prompt string,
	options ...contracts.GenerateOption,
) (<-chan contracts.StreamEvent, error) {
	// Apply options
	params := &contracts.GenerateOptions{
		LLMConfig: &contracts.LLMConfig{
			Temperature: 0.7,
		},
	}

	for _, option := range options {
		option(params)
	}

	// Check for organization ID in context
	defaultOrgID := "default"
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, defaultOrgID)
	}

	// Get buffer size from stream config
	bufferSize := 100
	if params.StreamConfig != nil {
		bufferSize = params.StreamConfig.BufferSize
	}

	// Create event channel
	eventChan := make(chan contracts.StreamEvent, bufferSize)

	// Start streaming in a goroutine
	go func() {
		defer close(eventChan)

		// Build messages
		messages := []Message{}

		// Add system message if provided
		if params.SystemMessage != "" {
			messages = append(messages, Message{
				Role:    "system",
				Content: params.SystemMessage,
			})
			c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
		}

		// Build messages using unified builder
		builder := deepSeekNewMessageHistoryBuilder(c.logger)
		messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)

		// Create stream request
		req := ChatCompletionRequest{
			Model:    c.Model,
			Messages: messages,
			Stream:   true,
		}

		if params.LLMConfig != nil {
			req.Temperature = params.LLMConfig.Temperature
			req.TopP = params.LLMConfig.TopP
			req.FrequencyPenalty = params.LLMConfig.FrequencyPenalty
			req.PresencePenalty = params.LLMConfig.PresencePenalty
			if len(params.LLMConfig.StopSequences) > 0 {
				req.Stop = params.LLMConfig.StopSequences
			}
		}

		// Add structured output if specified
		if params.ResponseFormat != nil {
			req.ResponseFormat = &ResponseFormatParam{
				Type:       "json_schema",
				JSONSchema: params.ResponseFormat.Schema,
			}
		}

		// Log the request
		c.logger.Debug(ctx, "Creating DeepSeek streaming request", map[string]interface{}{
			"model":       c.Model,
			"temperature": params.LLMConfig.Temperature,
			"top_p":       params.LLMConfig.TopP,
		})

		// Make streaming HTTP request
		resp, err := c.doStreamRequest(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "DeepSeek streaming error", map[string]interface{}{
				"error": err.Error(),
				"model": c.Model,
			})
			eventChan <- contracts.StreamEvent{
				Type:      contracts.StreamEventError,
				Error:     fmt.Errorf("deepseek streaming error: %w", err),
				Timestamp: time.Now(),
			}
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.Error(ctx, "Failed to close response body", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}()

		// Send initial message start event
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStart,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"model": c.Model,
			},
		}

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
				// Handle content delta
				if choice.Delta.Content != "" {
					eventChan <- contracts.StreamEvent{
						Type:      contracts.StreamEventContentDelta,
						Content:   choice.Delta.Content,
						Timestamp: time.Now(),
						Metadata: map[string]interface{}{
							"choice_index": choice.Index,
						},
					}
				}

				// Handle tool calls
				if len(choice.Delta.ToolCalls) > 0 {
					for _, toolCall := range choice.Delta.ToolCalls {
						if toolCall.Function.Name != "" || toolCall.Function.Arguments != "" {
							eventChan <- contracts.StreamEvent{
								Type: contracts.StreamEventToolUse,
								ToolCall: &contracts.ToolCall{
									ID:        toolCall.ID,
									Name:      toolCall.Function.Name,
									Arguments: toolCall.Function.Arguments,
								},
								Timestamp: time.Now(),
								Metadata: map[string]interface{}{
									"choice_index": choice.Index,
									"call_type":    "tool_call",
								},
							}
						}
					}
				}

				// Check for finish reason
				if choice.FinishReason != "" {
					eventChan <- contracts.StreamEvent{
						Type: contracts.StreamEventContentComplete,
						Metadata: map[string]interface{}{
							"finish_reason": choice.FinishReason,
							"choice_index":  choice.Index,
						},
						Timestamp: time.Now(),
					}
				}
			}

			// Handle usage information
			if chunk.Usage != nil && (chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0) {
				eventChan <- contracts.StreamEvent{
					Type:      contracts.StreamEventContentDelta,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"usage": map[string]interface{}{
							"prompt_tokens":     chunk.Usage.PromptTokens,
							"completion_tokens": chunk.Usage.CompletionTokens,
							"total_tokens":      chunk.Usage.TotalTokens,
						},
					},
				}
			}
		}

		// Check for scanner error
		if err := scanner.Err(); err != nil {
			c.logger.Error(ctx, "DeepSeek stream scanner error", map[string]interface{}{
				"error": err.Error(),
				"model": c.Model,
			})
			eventChan <- contracts.StreamEvent{
				Type:      contracts.StreamEventError,
				Error:     fmt.Errorf("deepseek stream scanner error: %w", err),
				Timestamp: time.Now(),
			}
			return
		}

		// Send final message stop event
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStop,
			Timestamp: time.Now(),
		}

		c.logger.Debug(ctx, "Successfully completed DeepSeek streaming request", map[string]interface{}{
			"model": c.Model,
		})
	}()

	return eventChan, nil
}

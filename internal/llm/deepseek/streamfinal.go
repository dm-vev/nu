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

func (c *Client) streamFinalSynthesis(
	ctx context.Context,
	params *contracts.GenerateOptions,
	messages []Message,
	maxIterations int,
	eventChan chan contracts.StreamEvent,
) {
	// Final call without tools to get synthesis
	c.logger.Info(ctx, "Maximum iterations reached, making final call without tools", map[string]interface{}{
		"maxIterations": maxIterations,
	})

	// Add explicit message to inform LLM this is the final call
	finalMessages := append(messages, Message{
		Role:    "user",
		Content: "Please provide your final response based on the information available. Do not request any additional tools.",
	})

	// Create final request without tools
	finalReq := ChatCompletionRequest{
		Model:    c.Model,
		Messages: finalMessages,
		Stream:   true,
	}

	if params.LLMConfig != nil {
		finalReq.Temperature = params.LLMConfig.Temperature
		finalReq.TopP = params.LLMConfig.TopP
		finalReq.FrequencyPenalty = params.LLMConfig.FrequencyPenalty
		finalReq.PresencePenalty = params.LLMConfig.PresencePenalty
	}

	// Add structured output if specified
	if params.ResponseFormat != nil {
		finalReq.ResponseFormat = &ResponseFormatParam{
			Type:       "json_schema",
			JSONSchema: params.ResponseFormat.Schema,
		}
	}

	c.logger.Debug(ctx, "Making final streaming call without tools", map[string]interface{}{
		"model": c.Model,
	})

	// Make final streaming request
	finalResp, err := c.doStreamRequest(ctx, finalReq)
	if err != nil {
		c.logger.Error(ctx, "Error in final streaming call without tools", map[string]interface{}{
			"error": err.Error(),
		})
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventError,
			Error:     fmt.Errorf("deepseek final streaming error: %w", err),
			Timestamp: time.Now(),
		}
		return
	}
	defer func() {
		if err := finalResp.Body.Close(); err != nil {
			c.logger.Error(ctx, "Failed to close final response body", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Process final stream
	finalScanner := bufio.NewScanner(finalResp.Body)
	for finalScanner.Scan() {
		line := finalScanner.Text()

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
			c.logger.Error(ctx, "Failed to parse final stream chunk", map[string]interface{}{
				"error": err.Error(),
				"data":  data,
			})
			continue
		}

		for _, choice := range chunk.Choices {
			// Handle final content
			if choice.Delta.Content != "" {
				eventChan <- contracts.StreamEvent{
					Type:      contracts.StreamEventContentDelta,
					Content:   choice.Delta.Content,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"choice_index": choice.Index,
						"final_call":   true,
					},
				}
			}

			// Check for finish reason
			if choice.FinishReason != "" {
				eventChan <- contracts.StreamEvent{
					Type: contracts.StreamEventContentComplete,
					Metadata: map[string]interface{}{
						"finish_reason": choice.FinishReason,
						"choice_index":  choice.Index,
						"final_call":    true,
					},
					Timestamp: time.Now(),
				}
			}
		}
	}

	// Check for final stream error
	if err := finalScanner.Err(); err != nil {
		c.logger.Error(ctx, "DeepSeek final streaming error", map[string]interface{}{
			"error": err.Error(),
			"model": c.Model,
		})
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventError,
			Error:     fmt.Errorf("deepseek final streaming error: %w", err),
			Timestamp: time.Now(),
		}
		return
	}

	// Send final message stop event
	eventChan <- contracts.StreamEvent{
		Type:      contracts.StreamEventMessageStop,
		Timestamp: time.Now(),
	}

	c.logger.Debug(ctx, "Successfully completed DeepSeek streaming request with tools", map[string]interface{}{
		"model": c.Model,
	})
}

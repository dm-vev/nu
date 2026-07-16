package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"nu/internal/contracts"
)

// executeBedrockStreaming handles streaming for AWS Bedrock using the AWS SDK
func (c *Client) executeBedrockStreaming(ctx context.Context, req *CompletionRequest, eventChan chan<- contracts.StreamEvent, cacheConfig *contracts.CacheConfig) error {
	c.logger.Debug(ctx, "Executing Bedrock streaming request", map[string]interface{}{
		"modelID": c.Model, "region": c.BedrockConfig.Region,
	})
	output, err := c.BedrockConfig.InvokeModelStream(ctx, c.Model, req, cacheConfig)
	if err != nil {
		return fmt.Errorf("failed to invoke Bedrock streaming: %w", err)
	}
	stream := output.GetStream()
	defer func() {
		if closeErr := stream.Close(); closeErr != nil {
			c.logger.Warn(ctx, "Failed to close Bedrock stream", map[string]interface{}{"error": closeErr.Error()})
		}
	}()

	thinkingBlocks := make(map[int]bool)
	toolBlocks := make(map[int]struct {
		ID        string
		Name      string
		InputJSON strings.Builder
	})
	for event := range stream.Events() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		switch e := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var rawEvent map[string]interface{}
			if err := json.Unmarshal(e.Value.Bytes, &rawEvent); err != nil {
				c.logger.Error(ctx, "Failed to parse Bedrock streaming chunk", map[string]interface{}{"error": err.Error()})
				continue
			}
			eventType, ok := rawEvent["type"].(string)
			if !ok {
				c.logger.Error(ctx, "Bedrock event missing type field", map[string]interface{}{"raw_event": string(e.Value.Bytes)})
				continue
			}
			anthropicEvent := &SSEEvent{Type: eventType, Data: e.Value.Bytes}
			streamEvent, err := c.convertAnthropicEventToStreamEvent(anthropicEvent, thinkingBlocks, toolBlocks)
			if err != nil {
				c.logger.Error(ctx, "Failed to convert Bedrock event", map[string]interface{}{"error": err.Error(), "event_type": eventType})
				continue
			}
			if streamEvent != nil {
				select {
				case eventChan <- *streamEvent:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		default:
			c.logger.Debug(ctx, "Unknown Bedrock streaming event type", map[string]interface{}{"type": fmt.Sprintf("%T", e)})
		}
	}
	if err := stream.Err(); err != nil {
		c.logger.Error(ctx, "Bedrock streaming error", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("bedrock streaming error: %w", err)
	}
	c.logger.Debug(ctx, "Successfully completed Bedrock streaming request", map[string]interface{}{"modelID": c.Model})
	return nil
}

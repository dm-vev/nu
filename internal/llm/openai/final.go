package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dm-vev/nu/contracts"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

func (c *Client) streamFinalToolResponse(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	params *contracts.GenerateOptions,
	maxIterations int,
	eventChan chan<- contracts.StreamEvent,
) {
	c.logger.Info(ctx, "Maximum iterations reached, making final call without tools", map[string]interface{}{"maxIterations": maxIterations})
	finalMessages := append(messages, openai.UserMessage("Please provide your final response based on the information available. Do not request any additional tools."))
	finalStreamParams := openai.ChatCompletionNewParams{Model: openai.ChatModel(c.Model), Messages: finalMessages}
	if !openAIIsReasoningModel(c.Model) {
		finalStreamParams.Temperature = openai.Float(params.LLMConfig.Temperature)
	}
	if params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
		finalStreamParams.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema},
		}
	}
	if params.LLMConfig != nil {
		if params.LLMConfig.TopP > 0 && !openAIIsReasoningModel(c.Model) {
			finalStreamParams.TopP = openai.Float(params.LLMConfig.TopP)
		}
		if params.LLMConfig.FrequencyPenalty != 0 {
			finalStreamParams.FrequencyPenalty = openai.Float(params.LLMConfig.FrequencyPenalty)
		}
		if params.LLMConfig.PresencePenalty != 0 {
			finalStreamParams.PresencePenalty = openai.Float(params.LLMConfig.PresencePenalty)
		}
		if openAIIsReasoningModel(c.Model) && params.LLMConfig.Reasoning != "" {
			finalStreamParams.ReasoningEffort = shared.ReasoningEffort(params.LLMConfig.Reasoning)
			c.logger.Debug(ctx, "Setting reasoning effort for final call", map[string]interface{}{"reasoning_effort": params.LLMConfig.Reasoning})
		}
	}
	c.logger.Debug(ctx, "Making final streaming call without tools", map[string]interface{}{"model": c.Model})
	finalStream := c.ChatService.Completions.NewStreaming(ctx, finalStreamParams)
	if finalStream.Err() != nil {
		c.logger.Error(ctx, "Error in final streaming call without tools", map[string]interface{}{"error": finalStream.Err().Error()})
		eventChan <- contracts.StreamEvent{
			Type: contracts.StreamEventError, Error: fmt.Errorf("openai final streaming error: %w", finalStream.Err()), Timestamp: time.Now(),
		}
		return
	}
	var finalContent strings.Builder
	for finalStream.Next() {
		chunk := finalStream.Current()
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				finalContent.WriteString(choice.Delta.Content)
				eventChan <- contracts.StreamEvent{
					Type: contracts.StreamEventContentDelta, Content: choice.Delta.Content, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"choice_index": choice.Index, "final_call": true},
				}
			}
			if choice.FinishReason != "" {
				eventChan <- contracts.StreamEvent{
					Type: contracts.StreamEventContentComplete, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"finish_reason": choice.FinishReason, "choice_index": choice.Index, "final_call": true},
				}
			}
		}
	}
	if err := finalStream.Err(); err != nil {
		c.logger.Error(ctx, "OpenAI final streaming error", map[string]interface{}{"error": err.Error(), "model": c.Model})
		eventChan <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("openai final streaming error: %w", err), Timestamp: time.Now()}
		return
	}
	eventChan <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
	c.logger.Debug(ctx, "Successfully completed OpenAI streaming request with tools", map[string]interface{}{"model": c.Model})
}

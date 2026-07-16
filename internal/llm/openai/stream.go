package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nu/internal/contracts"
	"nu/internal/multitenancy"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// GenerateStream implements contracts.StreamingLLM.GenerateStream.
func (c *Client) GenerateStream(ctx context.Context, prompt string, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, option := range options {
		option(params)
	}
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, "default")
	}
	bufferSize := 100
	if params.StreamConfig != nil {
		bufferSize = params.StreamConfig.BufferSize
	}
	eventChan := make(chan contracts.StreamEvent, bufferSize)

	go func() {
		defer close(eventChan)
		messages := []openai.ChatCompletionMessageParamUnion{}
		if params.SystemMessage != "" {
			messages = append(messages, openai.SystemMessage(params.SystemMessage))
			c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
		}
		builder := openAINewMessageHistoryBuilder(c.logger)
		messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)
		streamParams := openai.ChatCompletionNewParams{Model: openai.ChatModel(c.Model), Messages: messages}
		if !openAIIsReasoningModel(c.Model) {
			streamParams.Temperature = openai.Float(params.LLMConfig.Temperature)
		}
		if params.ResponseFormat != nil {
			jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
			streamParams.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema},
			}
		}
		if openAIIsReasoningModel(c.Model) || (params.LLMConfig != nil && params.LLMConfig.EnableReasoning) {
			streamParams.StreamOptions = openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)}
			if openAIIsReasoningModel(c.Model) {
				c.logger.Debug(ctx, "Using reasoning model with built-in reasoning", map[string]interface{}{
					"model": c.Model, "note": "reasoning models have internal reasoning but don't expose raw thinking tokens in streaming",
				})
			} else if params.LLMConfig != nil && params.LLMConfig.EnableReasoning {
				c.logger.Debug(ctx, "Reasoning enabled for non-reasoning model", map[string]interface{}{
					"model": c.Model, "note": "reasoning tokens not supported for this model type",
				})
			}
		}
		if params.LLMConfig != nil {
			if params.LLMConfig.TopP > 0 && !openAIIsReasoningModel(c.Model) {
				streamParams.TopP = openai.Float(params.LLMConfig.TopP)
			}
			if params.LLMConfig.FrequencyPenalty != 0 {
				streamParams.FrequencyPenalty = openai.Float(params.LLMConfig.FrequencyPenalty)
			}
			if params.LLMConfig.PresencePenalty != 0 {
				streamParams.PresencePenalty = openai.Float(params.LLMConfig.PresencePenalty)
			}
			if len(params.LLMConfig.StopSequences) > 0 {
				streamParams.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.LLMConfig.StopSequences}
			}
			if openAIIsReasoningModel(c.Model) && params.LLMConfig.Reasoning != "" {
				streamParams.ReasoningEffort = shared.ReasoningEffort(params.LLMConfig.Reasoning)
				c.logger.Debug(ctx, "Setting reasoning effort for streaming", map[string]interface{}{"reasoning_effort": params.LLMConfig.Reasoning})
			}
		}
		c.logger.Debug(ctx, "Creating OpenAI streaming request", map[string]interface{}{
			"model": c.Model, "temperature": params.LLMConfig.Temperature,
			"top_p": params.LLMConfig.TopP, "is_reasoning_model": openAIIsReasoningModel(c.Model),
		})
		stream := c.ChatService.Completions.NewStreaming(ctx, streamParams)
		eventChan <- contracts.StreamEvent{
			Type: contracts.StreamEventMessageStart, Timestamp: time.Now(), Metadata: map[string]interface{}{"model": c.Model},
		}
		var accumulatedContent strings.Builder
		for stream.Next() {
			chunk := stream.Current()
			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					accumulatedContent.WriteString(choice.Delta.Content)
					eventChan <- contracts.StreamEvent{
						Type: contracts.StreamEventContentDelta, Content: choice.Delta.Content, Timestamp: time.Now(),
						Metadata: map[string]interface{}{"choice_index": choice.Index},
					}
				}
				for _, toolCall := range choice.Delta.ToolCalls {
					if toolCall.Function.Name != "" || toolCall.Function.Arguments != "" {
						eventChan <- contracts.StreamEvent{
							Type:      contracts.StreamEventToolUse,
							ToolCall:  &contracts.ToolCall{ID: toolCall.ID, Name: toolCall.Function.Name, Arguments: toolCall.Function.Arguments},
							Timestamp: time.Now(), Metadata: map[string]interface{}{
								"choice_index": choice.Index, "call_type": "tool_call", "tool_index": toolCall.Index,
							},
						}
					}
				}
				if choice.FinishReason != "" {
					eventChan <- contracts.StreamEvent{
						Type: contracts.StreamEventContentComplete, Timestamp: time.Now(),
						Metadata: map[string]interface{}{"finish_reason": choice.FinishReason, "choice_index": choice.Index},
					}
				}
			}
			if chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 || chunk.Usage.TotalTokens > 0 {
				eventChan <- contracts.StreamEvent{
					Type: contracts.StreamEventContentDelta, Timestamp: time.Now(), Metadata: map[string]interface{}{
						"usage": map[string]interface{}{
							"prompt_tokens": chunk.Usage.PromptTokens, "completion_tokens": chunk.Usage.CompletionTokens,
							"total_tokens": chunk.Usage.TotalTokens,
						},
					},
				}
			}
		}
		if err := stream.Err(); err != nil {
			c.logger.Error(ctx, "OpenAI streaming error", map[string]interface{}{"error": err.Error(), "model": c.Model})
			eventChan <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("openai streaming error: %w", err), Timestamp: time.Now()}
			return
		}
		eventChan <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
		c.logger.Debug(ctx, "Successfully completed OpenAI streaming request", map[string]interface{}{"model": c.Model})
	}()
	return eventChan, nil
}

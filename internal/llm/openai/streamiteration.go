package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nu/internal/contracts"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

type openAIToolStreamIterationResult struct {
	messages      []openai.ChatCompletionMessageParamUnion
	contentEvents []contracts.StreamEvent
	hadContent    bool
	calledTools   bool
	complete      bool
	abort         bool
}

func (c *Client) runToolStreamIteration(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	openaiTools []openai.ChatCompletionToolUnionParam,
	tools []contracts.Tool,
	params *contracts.GenerateOptions,
	iteration, maxIterations int,
	filterIntermediateContent bool,
	eventChan chan<- contracts.StreamEvent,
) openAIToolStreamIterationResult {
	streamParams := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.Model), Messages: messages, Tools: openaiTools,
		ToolChoice: openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String("auto")},
	}
	if !openAIIsReasoningModel(c.Model) {
		streamParams.Temperature = openai.Float(params.LLMConfig.Temperature)
	}
	if openAIIsReasoningModel(c.Model) || (params.LLMConfig != nil && params.LLMConfig.EnableReasoning) {
		streamParams.StreamOptions = openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)}
		if openAIIsReasoningModel(c.Model) {
			c.logger.Debug(ctx, "Using reasoning model with built-in reasoning for tools", map[string]interface{}{
				"model": c.Model, "note": "reasoning models have internal reasoning but don't expose raw thinking tokens in streaming",
			})
		} else {
			c.logger.Debug(ctx, "Reasoning enabled for non-reasoning model with tools", map[string]interface{}{
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
		if openAIIsReasoningModel(c.Model) && params.LLMConfig.Reasoning != "" {
			streamParams.ReasoningEffort = shared.ReasoningEffort(params.LLMConfig.Reasoning)
			c.logger.Debug(ctx, "Setting reasoning effort for tools streaming", map[string]interface{}{"reasoning_effort": params.LLMConfig.Reasoning})
		}
	}
	c.logger.Debug(ctx, "Creating OpenAI streaming request with tools", map[string]interface{}{
		"model": c.Model, "tools": len(openaiTools), "temperature": params.LLMConfig.Temperature,
		"iteration": iteration + 1, "maxIterations": maxIterations, "message_count": len(messages),
	})
	if iteration > 0 {
		c.logger.Debug(ctx, "Messages array for iteration", map[string]interface{}{"iteration": iteration + 1, "message_count": len(messages)})
		for i, msg := range messages {
			c.logger.Debug(ctx, "Message details", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", msg)})
		}
	}

	stream := c.ChatService.Completions.NewStreaming(ctx, streamParams)
	if stream.Err() != nil {
		c.logger.Error(ctx, "Failed to create OpenAI streaming", map[string]interface{}{"error": stream.Err().Error()})
		eventChan <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("openai streaming error: %w", stream.Err()), Timestamp: time.Now()}
		return openAIToolStreamIterationResult{abort: true}
	}

	var currentToolCall *contracts.ToolCall
	var toolCallBuffer strings.Builder
	var assistantResponse openai.ChatCompletionMessage
	var contentEvents []contracts.StreamEvent
	hasContent := false
	for stream.Next() {
		chunk := stream.Current()
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				hasContent = true
				assistantResponse.Content += choice.Delta.Content
				contentEvent := contracts.StreamEvent{
					Type: contracts.StreamEventContentDelta, Content: choice.Delta.Content, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"choice_index": choice.Index, "iteration": iteration + 1},
				}
				if filterIntermediateContent && len(openaiTools) > 0 && iteration < maxIterations-1 {
					contentEvents = append(contentEvents, contentEvent)
				} else {
					eventChan <- contentEvent
				}
			}
			for _, toolCall := range choice.Delta.ToolCalls {
				if toolCall.Function.Name == "" && toolCall.Function.Arguments == "" {
					continue
				}
				if toolCall.Function.Name != "" {
					if currentToolCall != nil && toolCallBuffer.Len() > 0 {
						currentToolCall.Arguments = toolCallBuffer.String()
						eventChan <- contracts.StreamEvent{Type: contracts.StreamEventToolUse, ToolCall: currentToolCall, Timestamp: time.Now()}
					}
					currentToolCall = &contracts.ToolCall{ID: toolCall.ID, Name: toolCall.Function.Name}
					toolCallBuffer.Reset()
					assistantResponse.ToolCalls = append(assistantResponse.ToolCalls, openai.ChatCompletionMessageToolCallUnion{
						ID: toolCall.ID, Type: "function",
						Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: toolCall.Function.Name},
					})
					c.logger.Debug(ctx, "Started new tool call", map[string]interface{}{"tool_id": toolCall.ID, "tool_name": toolCall.Function.Name})
				}
				if toolCall.Function.Arguments != "" {
					toolCallBuffer.WriteString(toolCall.Function.Arguments)
					if len(assistantResponse.ToolCalls) > 0 {
						assistantResponse.ToolCalls[len(assistantResponse.ToolCalls)-1].Function.Arguments += toolCall.Function.Arguments
					}
				}
			}
			if choice.FinishReason == "tool_calls" && currentToolCall != nil {
				currentToolCall.Arguments = toolCallBuffer.String()
				eventChan <- contracts.StreamEvent{
					Type: contracts.StreamEventToolUse, ToolCall: currentToolCall, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"finish_reason": "tool_calls", "iteration": iteration + 1},
				}
				currentToolCall = nil
				toolCallBuffer.Reset()
				c.logger.Debug(ctx, "Finished tool calls", map[string]interface{}{"finish_reason": choice.FinishReason, "iteration": iteration + 1})
			}
		}
	}
	if err := stream.Err(); err != nil {
		c.logger.Error(ctx, "OpenAI streaming with tools error", map[string]interface{}{"error": err.Error(), "model": c.Model})
		eventChan <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("openai streaming error: %w", err), Timestamp: time.Now()}
		return openAIToolStreamIterationResult{abort: true}
	}
	if len(assistantResponse.ToolCalls) == 0 {
		if hasContent {
			eventChan <- contracts.StreamEvent{
				Type: contracts.StreamEventContentComplete, Timestamp: time.Now(), Metadata: map[string]interface{}{"iteration": iteration + 1},
			}
		}
		return openAIToolStreamIterationResult{messages: messages, complete: true}
	}

	c.logger.Info(ctx, "Processing tool calls", map[string]interface{}{"count": len(assistantResponse.ToolCalls), "iteration": iteration + 1})
	for i, tc := range assistantResponse.ToolCalls {
		c.logger.Debug(ctx, "Assistant tool call", map[string]interface{}{
			"index": i, "id": tc.ID, "id_length": len(tc.ID), "name": tc.Function.Name, "args_len": len(tc.Function.Arguments),
		})
	}
	assistantResponse.Role = "assistant"
	messages = append(messages, assistantResponse.ToParam())
	messages = c.executeStreamToolCalls(ctx, messages, assistantResponse.ToolCalls, tools, iteration, eventChan)
	return openAIToolStreamIterationResult{
		messages: messages, contentEvents: contentEvents, hadContent: hasContent, calledTools: true,
	}
}

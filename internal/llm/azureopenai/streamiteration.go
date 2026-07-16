package azureopenai

import (
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
	"nu/internal/contracts"
)

type azureOpenAIStreamIterationResult struct {
	assistant           openai.ChatCompletionMessage
	hasContent          bool
	iterationHasContent bool
	contentEvents       []contracts.StreamEvent
	failed              bool
}

func (c *Client) runToolStreamIteration(run *azureOpenAIStreamToolRun, iteration int) azureOpenAIStreamIterationResult {
	streamParams := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.deployment), Messages: run.messages, Tools: run.openaiTools,
		ToolChoice: openai.ChatCompletionToolChoiceOptionUnionParam{OfAuto: openai.String("auto")},
	}
	if !azureOpenAIIsReasoningModel(c.Model) {
		streamParams.Temperature = openai.Float(run.params.LLMConfig.Temperature)
	}
	if azureOpenAIIsReasoningModel(c.Model) || (run.params.LLMConfig != nil && run.params.LLMConfig.EnableReasoning) {
		streamParams.StreamOptions = openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)}
		if azureOpenAIIsReasoningModel(c.Model) {
			c.logger.Debug(run.ctx, "Using reasoning model with built-in reasoning for tools", map[string]interface{}{
				"model": c.Model, "deployment": c.deployment,
				"note": "reasoning models have internal reasoning but don't expose raw thinking tokens in streaming",
			})
		} else {
			c.logger.Debug(run.ctx, "Reasoning enabled for non-reasoning model with tools", map[string]interface{}{
				"model": c.Model, "deployment": c.deployment, "note": "reasoning tokens not supported for this model type",
			})
		}
	}
	if run.params.LLMConfig != nil {
		if run.params.LLMConfig.TopP > 0 && !azureOpenAIIsReasoningModel(c.Model) {
			streamParams.TopP = openai.Float(run.params.LLMConfig.TopP)
		}
		if run.params.LLMConfig.FrequencyPenalty != 0 {
			streamParams.FrequencyPenalty = openai.Float(run.params.LLMConfig.FrequencyPenalty)
		}
		if run.params.LLMConfig.PresencePenalty != 0 {
			streamParams.PresencePenalty = openai.Float(run.params.LLMConfig.PresencePenalty)
		}
		if azureOpenAIIsReasoningModel(c.Model) && run.params.LLMConfig.Reasoning != "" {
			streamParams.ReasoningEffort = shared.ReasoningEffort(run.params.LLMConfig.Reasoning)
			c.logger.Debug(run.ctx, "Setting reasoning effort for tools streaming", map[string]interface{}{"reasoning_effort": run.params.LLMConfig.Reasoning})
		}
	}
	c.logger.Debug(run.ctx, "Creating Azure OpenAI streaming request with tools", map[string]interface{}{
		"model": c.Model, "deployment": c.deployment, "tools": len(run.openaiTools),
		"temperature": run.params.LLMConfig.Temperature, "iteration": iteration + 1,
		"maxIterations": run.maxIterations, "message_count": len(run.messages),
	})
	if iteration > 0 {
		c.logger.Debug(run.ctx, "Messages array for iteration", map[string]interface{}{"iteration": iteration + 1, "message_count": len(run.messages)})
		for i, msg := range run.messages {
			c.logger.Debug(run.ctx, "Message details", map[string]interface{}{"index": i, "type": fmt.Sprintf("%T", msg)})
		}
	}

	stream := c.ChatService.Completions.NewStreaming(run.ctx, streamParams)
	if stream.Err() != nil {
		c.logger.Error(run.ctx, "Failed to create Azure OpenAI streaming", map[string]interface{}{"error": stream.Err().Error()})
		run.events <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("azure openai streaming error: %w", stream.Err()), Timestamp: time.Now()}
		return azureOpenAIStreamIterationResult{failed: true}
	}

	var result azureOpenAIStreamIterationResult
	var currentToolCall *contracts.ToolCall
	var toolCallBuffer strings.Builder
	for stream.Next() {
		chunk := stream.Current()
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				result.hasContent = true
				result.iterationHasContent = true
				result.assistant.Content += choice.Delta.Content
				contentEvent := contracts.StreamEvent{
					Type: contracts.StreamEventContentDelta, Content: choice.Delta.Content, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"choice_index": choice.Index, "iteration": iteration + 1},
				}
				if run.filterContent && len(run.openaiTools) > 0 && iteration < run.maxIterations-1 {
					result.contentEvents = append(result.contentEvents, contentEvent)
				} else {
					run.events <- contentEvent
				}
			}
			for _, toolCall := range choice.Delta.ToolCalls {
				if toolCall.Function.Name == "" && toolCall.Function.Arguments == "" {
					continue
				}
				if toolCall.Function.Name != "" {
					if currentToolCall != nil && toolCallBuffer.Len() > 0 {
						currentToolCall.Arguments = toolCallBuffer.String()
						run.events <- contracts.StreamEvent{Type: contracts.StreamEventToolUse, ToolCall: currentToolCall, Timestamp: time.Now()}
					}
					currentToolCall = &contracts.ToolCall{ID: toolCall.ID, Name: toolCall.Function.Name}
					toolCallBuffer.Reset()
					result.assistant.ToolCalls = append(result.assistant.ToolCalls, openai.ChatCompletionMessageToolCallUnion{
						ID: toolCall.ID, Type: "function",
						Function: openai.ChatCompletionMessageFunctionToolCallFunction{Name: toolCall.Function.Name},
					})
					c.logger.Debug(run.ctx, "Started new tool call", map[string]interface{}{"tool_id": toolCall.ID, "tool_name": toolCall.Function.Name})
				}
				if toolCall.Function.Arguments != "" {
					toolCallBuffer.WriteString(toolCall.Function.Arguments)
					if len(result.assistant.ToolCalls) > 0 {
						last := len(result.assistant.ToolCalls) - 1
						result.assistant.ToolCalls[last].Function.Arguments += toolCall.Function.Arguments
					}
				}
			}
			if choice.FinishReason == "tool_calls" && currentToolCall != nil {
				currentToolCall.Arguments = toolCallBuffer.String()
				run.events <- contracts.StreamEvent{
					Type: contracts.StreamEventToolUse, ToolCall: currentToolCall, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"finish_reason": "tool_calls", "iteration": iteration + 1},
				}
				currentToolCall = nil
				toolCallBuffer.Reset()
				c.logger.Debug(run.ctx, "Finished tool calls", map[string]interface{}{"finish_reason": choice.FinishReason, "iteration": iteration + 1})
			}
		}
	}
	if err := stream.Err(); err != nil {
		c.logger.Error(run.ctx, "Azure OpenAI streaming with tools error", map[string]interface{}{"error": err.Error(), "model": c.Model, "deployment": c.deployment})
		run.events <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("azure openai streaming error: %w", err), Timestamp: time.Now()}
		result.failed = true
	}
	return result
}

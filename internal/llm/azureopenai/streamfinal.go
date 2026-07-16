package azureopenai

import (
	"fmt"
	"strings"
	"time"

	"nu/internal/contracts"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

func (c *Client) runFinalToolStream(run *azureOpenAIStreamToolRun) {
	c.logger.Info(run.ctx, "Maximum iterations reached, making final call without tools", map[string]interface{}{"maxIterations": run.maxIterations})
	finalMessages := append(run.messages, openai.UserMessage("Please provide your final response based on the information available. Do not request any additional tools."))
	finalStreamParams := openai.ChatCompletionNewParams{Model: openai.ChatModel(c.deployment), Messages: finalMessages}
	if !azureOpenAIIsReasoningModel(c.Model) {
		finalStreamParams.Temperature = openai.Float(run.params.LLMConfig.Temperature)
	}
	if run.params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: run.params.ResponseFormat.Name, Schema: run.params.ResponseFormat.Schema}
		finalStreamParams.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema}}
	}
	if run.params.LLMConfig != nil {
		if run.params.LLMConfig.TopP > 0 && !azureOpenAIIsReasoningModel(c.Model) {
			finalStreamParams.TopP = openai.Float(run.params.LLMConfig.TopP)
		}
		if run.params.LLMConfig.FrequencyPenalty != 0 {
			finalStreamParams.FrequencyPenalty = openai.Float(run.params.LLMConfig.FrequencyPenalty)
		}
		if run.params.LLMConfig.PresencePenalty != 0 {
			finalStreamParams.PresencePenalty = openai.Float(run.params.LLMConfig.PresencePenalty)
		}
		if azureOpenAIIsReasoningModel(c.Model) && run.params.LLMConfig.Reasoning != "" {
			finalStreamParams.ReasoningEffort = shared.ReasoningEffort(run.params.LLMConfig.Reasoning)
			c.logger.Debug(run.ctx, "Setting reasoning effort for final call", map[string]interface{}{"reasoning_effort": run.params.LLMConfig.Reasoning})
		}
	}
	c.logger.Debug(run.ctx, "Making final streaming call without tools", map[string]interface{}{"model": c.Model, "deployment": c.deployment})
	finalStream := c.ChatService.Completions.NewStreaming(run.ctx, finalStreamParams)
	if finalStream.Err() != nil {
		c.logger.Error(run.ctx, "Error in final streaming call without tools", map[string]interface{}{"error": finalStream.Err().Error()})
		run.events <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("azure openai final streaming error: %w", finalStream.Err()), Timestamp: time.Now()}
		return
	}
	var finalContent strings.Builder
	for finalStream.Next() {
		chunk := finalStream.Current()
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				finalContent.WriteString(choice.Delta.Content)
				run.events <- contracts.StreamEvent{
					Type: contracts.StreamEventContentDelta, Content: choice.Delta.Content, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"choice_index": choice.Index, "final_call": true},
				}
			}
			if choice.FinishReason != "" {
				run.events <- contracts.StreamEvent{
					Type: contracts.StreamEventContentComplete, Timestamp: time.Now(),
					Metadata: map[string]interface{}{"finish_reason": choice.FinishReason, "choice_index": choice.Index, "final_call": true},
				}
			}
		}
	}
	if err := finalStream.Err(); err != nil {
		c.logger.Error(run.ctx, "Azure OpenAI final streaming error", map[string]interface{}{"error": err.Error(), "model": c.Model, "deployment": c.deployment})
		run.events <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: fmt.Errorf("azure openai final streaming error: %w", err), Timestamp: time.Now()}
		return
	}
	run.events <- contracts.StreamEvent{Type: contracts.StreamEventMessageStop, Timestamp: time.Now()}
	c.logger.Debug(run.ctx, "Successfully completed Azure OpenAI streaming request with tools", map[string]interface{}{"model": c.Model, "deployment": c.deployment})
}

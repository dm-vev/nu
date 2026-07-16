package openai

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/contracts"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

func (c *Client) newToolRequest(ctx context.Context, prompt string, tools []contracts.Tool, params *contracts.GenerateOptions) ([]openai.ChatCompletionMessageParamUnion, openai.ChatCompletionNewParams) {
	openaiTools := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {
		openaiTools[i] = openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name: tool.Name(), Description: openai.String(tool.Description()),
			Parameters: c.convertToOpenAISchema(tool.Parameters()),
		})
	}

	builder := openAINewMessageHistoryBuilder(c.logger)
	messages := builder.buildMessages(ctx, prompt, params.Memory)
	if params.SystemMessage != "" {
		messages = append(messages, openai.SystemMessage(params.SystemMessage))
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
	}

	req := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.Model), Messages: messages, Tools: openaiTools,
		Temperature: openai.Float(c.getTemperatureForModel(params.LLMConfig.Temperature)),
	}
	if params.LLMConfig.FrequencyPenalty != 0 {
		req.FrequencyPenalty = openai.Float(params.LLMConfig.FrequencyPenalty)
	}
	if params.LLMConfig.PresencePenalty != 0 {
		req.PresencePenalty = openai.Float(params.LLMConfig.PresencePenalty)
	}
	if !openAIIsReasoningModel(c.Model) && params.LLMConfig.TopP > 0 && params.LLMConfig.TopP <= 1 {
		req.TopP = openai.Float(params.LLMConfig.TopP)
	}
	if !openAIIsReasoningModel(c.Model) {
		req.ParallelToolCalls = openai.Bool(true)
	}
	if len(params.LLMConfig.StopSequences) > 0 {
		req.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.LLMConfig.StopSequences}
	}
	if openAIIsReasoningModel(c.Model) && params.LLMConfig.Reasoning != "" {
		req.ReasoningEffort = shared.ReasoningEffort(params.LLMConfig.Reasoning)
		c.logger.Debug(ctx, "Setting reasoning effort", map[string]interface{}{"reasoning_effort": params.LLMConfig.Reasoning})
	}
	if params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
		req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema},
		}
		c.logger.Debug(ctx, "Using response format", map[string]interface{}{"format": *params.ResponseFormat})
	}
	return messages, req
}

func (c *Client) finalToolResponse(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, params *contracts.GenerateOptions, maxIterations int) (string, error) {
	c.logger.Info(ctx, "Maximum iterations reached, making final call without tools", map[string]interface{}{"maxIterations": maxIterations})
	finalReq := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.Model), Messages: messages, Tools: nil,
		Temperature: openai.Float(c.getTemperatureForModel(params.LLMConfig.Temperature)),
	}
	if params.LLMConfig.FrequencyPenalty != 0 {
		finalReq.FrequencyPenalty = openai.Float(params.LLMConfig.FrequencyPenalty)
	}
	if params.LLMConfig.PresencePenalty != 0 {
		finalReq.PresencePenalty = openai.Float(params.LLMConfig.PresencePenalty)
	}
	if !openAIIsReasoningModel(c.Model) && params.LLMConfig.TopP > 0 && params.LLMConfig.TopP <= 1 {
		finalReq.TopP = openai.Float(params.LLMConfig.TopP)
	}
	if len(params.LLMConfig.StopSequences) > 0 {
		finalReq.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.LLMConfig.StopSequences}
	}
	if params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
		finalReq.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema},
		}
	}
	finalReq.Messages = append(finalReq.Messages, openai.SystemMessage("Please provide your final response based on the information available. Do not request any additional tools."))
	c.logger.Debug(ctx, "Making final request without tools", map[string]interface{}{"messages": len(finalReq.Messages)})
	finalResp, err := c.ChatService.Completions.New(ctx, finalReq)
	if err != nil {
		c.logger.Error(ctx, "Error in final call without tools", map[string]interface{}{"error": err.Error()})
		return "", fmt.Errorf("failed to create final chat completion: %w", err)
	}
	if len(finalResp.Choices) == 0 {
		return "", fmt.Errorf("no completions returned in final call")
	}
	if acc := openAIGetUsageAccumulator(ctx); acc != nil {
		acc.add(int(finalResp.Usage.PromptTokens), int(finalResp.Usage.CompletionTokens), int(finalResp.Usage.TotalTokens),
			int(finalResp.Usage.CompletionTokensDetails.ReasoningTokens), c.Model)
	}
	content := strings.TrimSpace(finalResp.Choices[0].Message.Content)
	c.logger.Info(ctx, "Successfully received final response without tools", nil)
	return content, nil
}

package azureopenai

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/contracts"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

func (c *Client) generateFinalToolResponse(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	params *contracts.GenerateOptions,
	maxIterations int,
) (string, error) {
	c.logger.Info(ctx, "Maximum iterations reached, making final call without tools", map[string]interface{}{"maxIterations": maxIterations})
	finalReq := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.deployment), Messages: messages, Tools: nil,
		Temperature:      openai.Float(c.getTemperatureForModel(params.LLMConfig.Temperature)),
		FrequencyPenalty: openai.Float(params.LLMConfig.FrequencyPenalty),
		PresencePenalty:  openai.Float(params.LLMConfig.PresencePenalty),
	}
	if !azureOpenAIIsReasoningModel(c.Model) {
		finalReq.TopP = openai.Float(params.LLMConfig.TopP)
	}
	if len(params.LLMConfig.StopSequences) > 0 {
		finalReq.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.LLMConfig.StopSequences}
	}
	if params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
		finalReq.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema}}
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
	content := strings.TrimSpace(finalResp.Choices[0].Message.Content)
	c.logger.Info(ctx, "Successfully received final response without tools", nil)
	return content, nil
}

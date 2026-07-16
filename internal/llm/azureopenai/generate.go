package azureopenai

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// Generate generates text from a prompt
func (c *Client) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	response, err := c.generateInternal(ctx, prompt, options...)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// generateInternal performs the actual generation and returns the full response
func (c *Client) generateInternal(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, option := range options {
		option(params)
	}

	orgID, _ := multitenancy.GetOrgID(ctx)
	if orgID != "" {
		ctx = context.WithValue(ctx, azureOpenAIOrganizationKey, orgID)
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	if params.SystemMessage != "" {
		messages = append(messages, openai.SystemMessage(params.SystemMessage))
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
	}
	builder := azureOpenAINewMessageHistoryBuilder(c.logger)
	messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)

	req := openai.ChatCompletionNewParams{Model: openai.ChatModel(c.deployment), Messages: messages}
	if params.LLMConfig != nil {
		req.Temperature = openai.Float(c.getTemperatureForModel(params.LLMConfig.Temperature))
		if !azureOpenAIIsReasoningModel(c.Model) {
			req.TopP = openai.Float(params.LLMConfig.TopP)
		}
		req.FrequencyPenalty = openai.Float(params.LLMConfig.FrequencyPenalty)
		req.PresencePenalty = openai.Float(params.LLMConfig.PresencePenalty)
		if len(params.LLMConfig.StopSequences) > 0 {
			req.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.LLMConfig.StopSequences}
		}
		if azureOpenAIIsReasoningModel(c.Model) && params.LLMConfig.Reasoning != "" {
			req.ReasoningEffort = shared.ReasoningEffort(params.LLMConfig.Reasoning)
			c.logger.Debug(ctx, "Setting reasoning effort", map[string]interface{}{"reasoning_effort": params.LLMConfig.Reasoning})
		}
	}

	if params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
		req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema}}
		c.logger.Debug(ctx, "Using response format", map[string]interface{}{"format": *params.ResponseFormat})
	}
	if orgID, ok := ctx.Value(azureOpenAIOrganizationKey).(string); ok && orgID != "" {
		req.User = openai.String(orgID)
	}

	var resp *openai.ChatCompletion
	var err error
	operation := func() error {
		reasoningEffort := "none"
		if params.LLMConfig != nil && params.LLMConfig.Reasoning != "" {
			reasoningEffort = params.LLMConfig.Reasoning
		}
		c.logger.Debug(ctx, "Executing Azure OpenAI API request", map[string]interface{}{
			"model": c.Model, "deployment": c.deployment, "temperature": req.Temperature,
			"top_p": req.TopP, "frequency_penalty": req.FrequencyPenalty,
			"presence_penalty": req.PresencePenalty, "stop_sequences": req.Stop,
			"messages": len(req.Messages), "response_format": params.ResponseFormat != nil,
			"reasoning_effort": reasoningEffort,
		})
		resp, err = c.ChatService.Completions.New(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from Azure OpenAI API", map[string]interface{}{"error": err.Error(), "model": c.Model, "deployment": c.deployment})
			return fmt.Errorf("failed to generate text: %w", err)
		}
		return nil
	}

	if c.retryExecutor != nil {
		c.logger.Debug(ctx, "Using retry mechanism for Azure OpenAI request", map[string]interface{}{"model": c.Model, "deployment": c.deployment})
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		err = operation()
	}
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Azure OpenAI API")
	}

	c.logger.Debug(ctx, "Successfully received response from Azure OpenAI", map[string]interface{}{"model": c.Model, "deployment": c.deployment})
	response := &contracts.LLMResponse{
		Content: resp.Choices[0].Message.Content, Model: string(resp.Model),
		StopReason: string(resp.Choices[0].FinishReason),
		Metadata:   map[string]interface{}{"provider": "azure_openai", "deployment": c.deployment},
	}
	usage := &contracts.TokenUsage{
		InputTokens: int(resp.Usage.PromptTokens), OutputTokens: int(resp.Usage.CompletionTokens), TotalTokens: int(resp.Usage.TotalTokens),
	}
	if resp.Usage.CompletionTokensDetails.ReasoningTokens > 0 {
		usage.ReasoningTokens = int(resp.Usage.CompletionTokensDetails.ReasoningTokens)
	}
	response.Usage = usage
	return response, nil
}

// GenerateDetailed generates text and returns detailed response information including token usage
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	return c.generateInternal(ctx, prompt, options...)
}

package openai

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// Generate generates text from a prompt.
func (c *Client) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	response, err := c.generateInternal(ctx, prompt, options...)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// GenerateDetailed generates text and returns token usage and response metadata.
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	return c.generateInternal(ctx, prompt, options...)
}

func (c *Client) generateInternal(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, option := range options {
		option(params)
	}

	orgID, _ := multitenancy.GetOrgID(ctx)
	if orgID != "" {
		ctx = context.WithValue(ctx, openAIOrganizationKey, orgID)
	}

	messages := []openai.ChatCompletionMessageParamUnion{}
	if params.SystemMessage != "" {
		messages = append(messages, openai.SystemMessage(params.SystemMessage))
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
	}
	builder := openAINewMessageHistoryBuilder(c.logger)
	messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)

	req := openai.ChatCompletionNewParams{Model: openai.ChatModel(c.Model), Messages: messages}
	if params.LLMConfig != nil {
		req.Temperature = openai.Float(c.getTemperatureForModel(params.LLMConfig.Temperature))
		if !openAIIsReasoningModel(c.Model) && params.LLMConfig.TopP > 0 && params.LLMConfig.TopP <= 1 {
			req.TopP = openai.Float(params.LLMConfig.TopP)
		}
		if params.LLMConfig.FrequencyPenalty != 0 {
			req.FrequencyPenalty = openai.Float(params.LLMConfig.FrequencyPenalty)
		}
		if params.LLMConfig.PresencePenalty != 0 {
			req.PresencePenalty = openai.Float(params.LLMConfig.PresencePenalty)
		}
		if len(params.LLMConfig.StopSequences) > 0 {
			req.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: params.LLMConfig.StopSequences}
		}
		if openAIIsReasoningModel(c.Model) && params.LLMConfig.Reasoning != "" {
			req.ReasoningEffort = shared.ReasoningEffort(params.LLMConfig.Reasoning)
			c.logger.Debug(ctx, "Setting reasoning effort", map[string]interface{}{"reasoning_effort": params.LLMConfig.Reasoning})
		}
	}

	if params.ResponseFormat != nil {
		jsonSchema := shared.ResponseFormatJSONSchemaJSONSchemaParam{Name: params.ResponseFormat.Name, Schema: params.ResponseFormat.Schema}
		req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{Type: "json_schema", JSONSchema: jsonSchema},
		}
		c.logger.Debug(ctx, "Using response format", map[string]interface{}{"format": *params.ResponseFormat})
	}
	if orgID, ok := ctx.Value(openAIOrganizationKey).(string); ok && orgID != "" {
		req.User = openai.String(orgID)
	}

	var resp *openai.ChatCompletion
	var err error
	operation := func() error {
		reasoningEffort := "none"
		if params.LLMConfig != nil && params.LLMConfig.Reasoning != "" {
			reasoningEffort = params.LLMConfig.Reasoning
		}
		c.logger.Debug(ctx, "Executing OpenAI API request", map[string]interface{}{
			"model": c.Model, "temperature": req.Temperature, "top_p": req.TopP,
			"frequency_penalty": req.FrequencyPenalty, "presence_penalty": req.PresencePenalty,
			"stop_sequences": req.Stop, "messages": len(req.Messages),
			"response_format": params.ResponseFormat != nil, "reasoning_effort": reasoningEffort,
		})
		resp, err = c.ChatService.Completions.New(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from OpenAI API", map[string]interface{}{"error": err.Error(), "model": c.Model})
			return fmt.Errorf("failed to generate text: %w", err)
		}
		return nil
	}

	if c.retryExecutor != nil {
		c.logger.Debug(ctx, "Using retry mechanism for OpenAI request", map[string]interface{}{"model": c.Model})
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		err = operation()
	}
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI API")
	}
	c.logger.Debug(ctx, "Successfully received response from OpenAI", map[string]interface{}{"model": c.Model})
	response := &contracts.LLMResponse{
		Content: resp.Choices[0].Message.Content, Model: string(resp.Model),
		StopReason: string(resp.Choices[0].FinishReason),
		Metadata:   map[string]interface{}{"provider": "openai"},
		Usage: &contracts.TokenUsage{
			InputTokens: int(resp.Usage.PromptTokens), OutputTokens: int(resp.Usage.CompletionTokens),
			TotalTokens: int(resp.Usage.TotalTokens),
		},
	}
	if resp.Usage.CompletionTokensDetails.ReasoningTokens > 0 {
		response.Usage.ReasoningTokens = int(resp.Usage.CompletionTokensDetails.ReasoningTokens)
	}
	return response, nil
}

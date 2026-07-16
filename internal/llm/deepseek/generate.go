package deepseek

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// Generate generates text based on the provided prompt
func (c *Client) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	response, err := c.GenerateDetailed(ctx, prompt, options...)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// GenerateDetailed generates text and returns detailed response information including token usage
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	// Apply options
	params := &contracts.GenerateOptions{
		LLMConfig: &contracts.LLMConfig{
			Temperature: 0.7,
		},
	}

	for _, option := range options {
		option(params)
	}

	// Get organization ID from context if available
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Build messages
	messages := []Message{}

	// Add system message if available
	if params.SystemMessage != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: params.SystemMessage,
		})
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
	}

	// Build messages using message history builder
	builder := deepSeekNewMessageHistoryBuilder(c.logger)
	messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)

	// Create request
	req := ChatCompletionRequest{
		Model:    c.Model,
		Messages: messages,
	}

	if params.LLMConfig != nil {
		req.Temperature = params.LLMConfig.Temperature
		req.TopP = params.LLMConfig.TopP
		req.FrequencyPenalty = params.LLMConfig.FrequencyPenalty
		req.PresencePenalty = params.LLMConfig.PresencePenalty
		if len(params.LLMConfig.StopSequences) > 0 {
			req.Stop = params.LLMConfig.StopSequences
		}
	}

	// Set response format if provided
	if params.ResponseFormat != nil {
		req.ResponseFormat = &ResponseFormatParam{
			Type:       "json_schema",
			JSONSchema: params.ResponseFormat.Schema,
		}
		c.logger.Debug(ctx, "Using response format", map[string]interface{}{"format": params.ResponseFormat})
	}

	var resp *ChatCompletionResponse
	var err error

	operation := func() error {
		c.logger.Debug(ctx, "Executing DeepSeek API request", map[string]interface{}{
			"model":             c.Model,
			"temperature":       req.Temperature,
			"top_p":             req.TopP,
			"frequency_penalty": req.FrequencyPenalty,
			"presence_penalty":  req.PresencePenalty,
			"stop_sequences":    req.Stop,
			"messages":          len(req.Messages),
			"response_format":   params.ResponseFormat != nil,
			"org_id":            orgID,
		})

		resp, err = c.doRequest(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from DeepSeek API", map[string]interface{}{
				"error": err.Error(),
				"model": c.Model,
			})
			return fmt.Errorf("failed to generate text: %w", err)
		}
		return nil
	}

	if c.retryExecutor != nil {
		c.logger.Debug(ctx, "Using retry mechanism for DeepSeek request", map[string]interface{}{
			"model": c.Model,
		})
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		err = operation()
	}

	if err != nil {
		return nil, err
	}

	// Return response
	if len(resp.Choices) > 0 {
		c.logger.Debug(ctx, "Successfully received response from DeepSeek", map[string]interface{}{
			"model": c.Model,
		})

		// Create detailed response with token usage
		response := &contracts.LLMResponse{
			Content:    resp.Choices[0].Message.Content,
			Model:      resp.Model,
			StopReason: resp.Choices[0].FinishReason,
			Metadata: map[string]interface{}{
				"provider": "deepseek",
			},
		}

		// Extract token usage
		response.Usage = &contracts.TokenUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		}

		return response, nil
	}

	return nil, fmt.Errorf("no response from DeepSeek API")
}

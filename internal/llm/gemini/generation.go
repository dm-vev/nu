package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
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

	// Build contents with memory and current prompt
	contents := c.buildContentsWithMemory(ctx, prompt, params)

	// Add system instruction if provided or if reasoning is specified
	var systemInstruction *genai.Content
	systemMessage := params.SystemMessage

	// Log reasoning mode usage - only affects native thinking models (2.5 series)
	if params.LLMConfig != nil && params.LLMConfig.Reasoning != "" {
		if SupportsThinking(c.model) {
			c.logger.Debug(ctx, "Using reasoning mode with thinking-capable model", map[string]interface{}{
				"reasoning": params.LLMConfig.Reasoning,
				"model":     c.model,
			})
		} else {
			c.logger.Debug(ctx, "Reasoning mode specified for non-thinking model - native thinking tokens not available", map[string]interface{}{
				"reasoning":        params.LLMConfig.Reasoning,
				"model":            c.model,
				"supportsThinking": false,
			})
		}
	}

	if systemMessage != "" {
		systemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: systemMessage},
			},
		}
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": systemMessage})
	}

	// Set generation config
	var genConfig *genai.GenerationConfig
	if params.LLMConfig != nil {
		genConfig = &genai.GenerationConfig{}

		if params.LLMConfig.Temperature > 0 {
			temp := float32(params.LLMConfig.Temperature)
			genConfig.Temperature = &temp
		}
		if params.LLMConfig.TopP > 0 {
			topP := float32(params.LLMConfig.TopP)
			genConfig.TopP = &topP
		}
		if len(params.LLMConfig.StopSequences) > 0 {
			genConfig.StopSequences = params.LLMConfig.StopSequences
		}
	}

	// Apply max output tokens if configured at client level
	c.applyMaxOutputTokens(&genConfig)

	// Set response format if provided
	if params.ResponseFormat != nil {
		if genConfig == nil {
			genConfig = &genai.GenerationConfig{}
		}

		genConfig.ResponseMIMEType = "application/json"

		// Convert schema for genai
		if schemaBytes, err := json.Marshal(params.ResponseFormat.Schema); err == nil {
			var schema *genai.Schema
			if err := json.Unmarshal(schemaBytes, &schema); err != nil {
				c.logger.Warn(ctx, "Failed to convert response schema", map[string]interface{}{"error": err.Error()})
			} else {
				genConfig.ResponseSchema = schema
			}
		}
		c.logger.Debug(ctx, "Using response format", map[string]interface{}{"format": *params.ResponseFormat})
	}

	var result *genai.GenerateContentResponse
	var err error

	operation := func() error {
		c.logger.Debug(ctx, "Executing Gemini API request", map[string]interface{}{
			"model":           c.model,
			"temperature":     genConfig.Temperature,
			"top_p":           genConfig.TopP,
			"stop_sequences":  genConfig.StopSequences,
			"response_format": params.ResponseFormat != nil,
			"org_id":          orgID,
		})

		config := &genai.GenerateContentConfig{
			SystemInstruction: systemInstruction,
		}

		// Apply generation config parameters directly to config
		if genConfig != nil {
			if genConfig.Temperature != nil {
				config.Temperature = genConfig.Temperature
			}
			if genConfig.TopP != nil {
				config.TopP = genConfig.TopP
			}
			if len(genConfig.StopSequences) > 0 {
				config.StopSequences = genConfig.StopSequences
			}
			if genConfig.ResponseMIMEType != "" {
				config.ResponseMIMEType = genConfig.ResponseMIMEType
			}
			if genConfig.ResponseSchema != nil {
				config.ResponseSchema = genConfig.ResponseSchema
			}
		}

		// Add thinking configuration if supported and enabled
		if SupportsThinking(c.model) && c.thinkingConfig != nil {
			if c.thinkingConfig.IncludeThoughts || c.thinkingConfig.ThinkingBudget != nil {
				config.ThinkingConfig = &genai.ThinkingConfig{
					IncludeThoughts: c.thinkingConfig.IncludeThoughts,
					ThinkingBudget:  c.thinkingConfig.ThinkingBudget,
				}

				c.logger.Debug(ctx, "Enabled thinking configuration", map[string]interface{}{
					"includeThoughts": c.thinkingConfig.IncludeThoughts,
					"thinkingBudget":  c.thinkingConfig.ThinkingBudget,
				})
			}
		}

		result, err = c.genaiClient.Models.GenerateContent(ctx, c.model, contents, config)
		if err != nil {
			c.logger.Error(ctx, "Error from Gemini API", map[string]interface{}{
				"error": err.Error(),
				"model": c.model,
			})
			return fmt.Errorf("failed to generate text: %w", err)
		}
		return nil
	}

	if c.retryExecutor != nil {
		c.logger.Debug(ctx, "Using retry mechanism for Gemini request", map[string]interface{}{
			"model": c.model,
		})
		err = c.retryExecutor.Execute(ctx, operation)
	} else {
		err = operation()
	}

	if err != nil {
		return nil, err
	}

	// Extract response and separate thinking from final content
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		c.logger.Debug(ctx, "Successfully received response from Gemini", map[string]interface{}{
			"model": c.model,
		})

		var textParts []string
		var thinkingParts []string

		for _, part := range result.Candidates[0].Content.Parts {
			if part.Text != "" {
				if part.Thought {
					// This is thinking content
					thinkingParts = append(thinkingParts, part.Text)
					c.logger.Debug(ctx, "Received thinking content", map[string]interface{}{
						"length": len(part.Text),
					})
				} else {
					// This is final response content
					textParts = append(textParts, part.Text)
				}
			}
		}

		// For non-streaming Generate, we return only the final response content
		// The thinking content is available but not returned in this interface
		// (it would be available in streaming through StreamEventThinking)
		if len(thinkingParts) > 0 {
			c.logger.Info(ctx, "Thinking content received but not included in response", map[string]interface{}{
				"thinkingParts": len(thinkingParts),
				"finalParts":    len(textParts),
			})
		}

		content := strings.Join(textParts, "")

		// Create detailed response with token usage
		response := &contracts.LLMResponse{
			Content:    content,
			Model:      c.model,
			StopReason: "", // Gemini doesn't provide specific stop reason
			Metadata: map[string]interface{}{
				"provider": "gemini",
			},
		}

		// Extract token usage if available
		if result.UsageMetadata != nil {
			usage := &contracts.TokenUsage{
				InputTokens:  int(result.UsageMetadata.PromptTokenCount),
				OutputTokens: int(result.UsageMetadata.CandidatesTokenCount),
				TotalTokens:  int(result.UsageMetadata.TotalTokenCount),
			}

			// Add thinking tokens if available (for 2.5 series models)
			// Note: Thinking token count may not be directly available in current genai library version
			if len(thinkingParts) > 0 {
				// For now, we note that thinking tokens were used but don't have the exact count
				// This will be updated when the genai library supports thinking token counting
				c.logger.Debug(ctx, "Thinking tokens used but count not available", map[string]interface{}{
					"thinkingParts": len(thinkingParts),
				})
			}

			response.Usage = usage
		}

		return response, nil
	}

	return nil, fmt.Errorf("no response from Gemini API")
}

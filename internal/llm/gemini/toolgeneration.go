package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/genai"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// GenerateWithTools implements contracts.LLM.GenerateWithTools
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	// Convert options to params
	params := &contracts.GenerateOptions{}
	for _, opt := range options {
		if opt != nil {
			opt(params)
		}
	}

	// Set default values only if they're not provided
	if params.LLMConfig == nil {
		params.LLMConfig = &contracts.LLMConfig{
			Temperature:      0.7,
			TopP:             1.0,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
		}
	}

	// Set default max iterations if not provided
	maxIterations := params.MaxIterations
	if maxIterations == 0 {
		maxIterations = 2 // Default to current behavior
	}

	// Check for organization ID in context
	orgID := "default"
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		orgID = id
	}
	_ = orgID // Mark as used to avoid linter warning

	// Convert tools to Gemini format. Shared with GenerateStreamWithTools so
	// Ask and Stream agree on schema shape, including array `items`.
	geminiTools := geminiConvertToolsToFunctionDeclarations(tools)

	// Build contents with memory and current prompt
	contents := c.buildContentsWithMemory(ctx, prompt, params)
	var systemInstruction *genai.Content

	// Track tool call repetitions for loop detection
	toolCallHistory := make(map[string]int)
	var toolCallHistoryMu sync.Mutex

	// Add system message if available
	if params.SystemMessage != "" {
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

		systemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: systemMessage},
			},
		}
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": systemMessage})
	}

	// Iterative tool calling loop
	// Track the last response content from the tool-calling loop
	var lastContent string

	for iteration := 0; iteration < maxIterations; iteration++ {
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

		logData := map[string]interface{}{
			"model":           c.model,
			"contents":        len(contents),
			"tools":           len(geminiTools),
			"response_format": params.ResponseFormat != nil,
			"iteration":       iteration + 1,
			"maxIterations":   maxIterations,
		}

		if genConfig != nil {
			if genConfig.Temperature != nil {
				logData["temperature"] = *genConfig.Temperature
			}
			if genConfig.TopP != nil {
				logData["top_p"] = *genConfig.TopP
			}
			if len(genConfig.StopSequences) > 0 {
				logData["stop_sequences"] = genConfig.StopSequences
			}
		}

		c.logger.Debug(ctx, "Sending request with tools to Gemini", logData)

		config := &genai.GenerateContentConfig{
			Tools: []*genai.Tool{
				{
					FunctionDeclarations: geminiTools,
				},
			},
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

		result, err := c.genaiClient.Models.GenerateContent(ctx, c.model, contents, config)
		if err != nil {
			c.logger.Error(ctx, "Error from Gemini API", map[string]interface{}{"error": err.Error()})
			return "", fmt.Errorf("failed to create content: %w", err)
		}

		if len(result.Candidates) == 0 {
			return "", fmt.Errorf("no candidates returned")
		}

		candidate := result.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			return "", fmt.Errorf("no content in response")
		}

		// Check if any part geminiContains function calls
		hasFunctionCalls := false
		for _, part := range candidate.Content.Parts {
			if part.FunctionCall != nil {
				hasFunctionCalls = true
				break
			}
		}

		// Extract text content from this iteration
		var iterTextParts []string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				iterTextParts = append(iterTextParts, part.Text)
			}
		}
		if len(iterTextParts) > 0 {
			lastContent = strings.Join(iterTextParts, " ")
		}

		// If no function calls, return the text response
		if !hasFunctionCalls {
			return lastContent, nil
		}

		contents, err = c.executeToolCalls(ctx, contents, candidate.Content.Parts, tools, params, toolCallHistory, &toolCallHistoryMu, iteration)
		if err != nil {
			return "", err
		}
	}

	return c.generateFinalToolResponse(ctx, contents, systemInstruction, params, maxIterations, lastContent)
}

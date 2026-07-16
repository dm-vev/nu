package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"nu/internal/contracts"
)

func (c *Client) generateFinalToolResponse(
	ctx context.Context,
	contents []*genai.Content,
	systemInstruction *genai.Content,
	params *contracts.GenerateOptions,
	maxIterations int,
	lastContent string,
) (string, error) {
	// If DisableFinalSummary is enabled, return the last response from the tool-calling loop
	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final summary call", map[string]interface{}{
			"maxIterations": maxIterations,
		})
		return lastContent, nil
	}

	c.logger.Info(ctx, "Maximum iterations reached, making final call without tools", map[string]interface{}{
		"maxIterations": maxIterations,
	})

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
	}

	// Add a conclusion instruction to the contents
	contents = append(contents, &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: "Please provide your final response based on the information available. Do not request any additional functions."},
		},
	})

	c.logger.Debug(ctx, "Making final request without tools", map[string]interface{}{
		"contents": len(contents),
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

	finalResult, err := c.genaiClient.Models.GenerateContent(ctx, c.model, contents, config)
	if err != nil {
		c.logger.Error(ctx, "Error in final call without tools", map[string]interface{}{"error": err.Error()})
		return "", fmt.Errorf("failed to create final content: %w", err)
	}

	if len(finalResult.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned in final call")
	}

	candidate := finalResult.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content in final response")
	}

	// Extract text from all parts
	var textParts []string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			textParts = append(textParts, part.Text)
		}
	}

	content := strings.TrimSpace(strings.Join(textParts, " "))
	c.logger.Info(ctx, "Successfully received final response without tools", nil)
	return content, nil
}

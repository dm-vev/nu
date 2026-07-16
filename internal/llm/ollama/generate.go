package ollama

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dm-vev/nu/contracts"
)

// Generate generates text from a prompt
func (c *Client) Generate(ctx context.Context, prompt string, options ...contracts.GenerateOption) (string, error) {
	// Apply options
	params := &contracts.GenerateOptions{
		LLMConfig: &contracts.LLMConfig{
			Temperature: 0.7,
		},
	}

	for _, option := range options {
		option(params)
	}

	// Build prompt with memory context
	finalPrompt := c.buildPromptWithMemory(ctx, prompt, params)

	// Create request
	req := GenerateRequest{
		Model:  c.Model,
		Prompt: finalPrompt,
		Stream: false,
		Options: &Options{
			Temperature: params.LLMConfig.Temperature,
			TopP:        params.LLMConfig.TopP,
			Stop:        params.LLMConfig.StopSequences,
		},
		System: params.SystemMessage,
	}

	// Handle structured output if provided
	if params.ResponseFormat != nil && params.ResponseFormat.Type == contracts.ResponseFormatJSON {
		// Add JSON schema to the prompt for Ollama
		schemaJSON, err := json.Marshal(params.ResponseFormat.Schema)
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON schema: %w", err)
		}

		schemaPrompt := fmt.Sprintf(`%s

Please respond with a valid JSON object that matches the following schema:

Schema Name: %s
JSON Schema: %s

Ensure your response is a valid JSON object that strictly follows the schema above.`,
			prompt,
			params.ResponseFormat.Name,
			string(schemaJSON))

		req.Prompt = schemaPrompt
		req.Format = "json"
	}

	// Make request
	resp, err := c.makeRequest(ctx, "/api/generate", req)
	if err != nil {
		return "", fmt.Errorf("failed to generate text: %w", err)
	}

	var generateResp GenerateResponse
	if err := json.Unmarshal(resp, &generateResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return generateResp.Response, nil
}

// GenerateDetailed generates text and returns detailed response information including token usage
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	// Call the existing method and construct a detailed response
	content, err := c.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}

	// Return a detailed response without usage information (Ollama doesn't provide token usage)
	return &contracts.LLMResponse{
		Content:    content,
		Model:      c.Model,
		StopReason: "",
		Usage:      nil, // Ollama doesn't provide token usage information
		Metadata: map[string]interface{}{
			"provider": "ollama",
		},
	}, nil
}

// GenerateWithToolsDetailed generates text with tools and returns detailed response information including token usage
func (c *Client) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	// Call the existing method and construct a detailed response
	content, err := c.GenerateWithTools(ctx, prompt, tools, options...)
	if err != nil {
		return nil, err
	}

	// Return a detailed response without usage information
	return &contracts.LLMResponse{
		Content:    content,
		Model:      c.Model,
		StopReason: "",
		Usage:      nil, // Ollama doesn't provide token usage information
		Metadata: map[string]interface{}{
			"provider":   "ollama",
			"tools_used": true,
		},
	}, nil
}

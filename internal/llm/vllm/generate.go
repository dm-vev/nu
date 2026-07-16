package vllm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
		Model:       c.Model,
		Prompt:      finalPrompt,
		Stream:      false,
		Temperature: params.LLMConfig.Temperature,
		TopP:        params.LLMConfig.TopP,
		Stop:        params.LLMConfig.StopSequences,
	}

	// Handle structured output if provided
	if params.ResponseFormat != nil && params.ResponseFormat.Type == contracts.ResponseFormatJSON {
		// Add JSON schema to the prompt for vLLM
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
	}

	// Make request
	resp, err := c.makeRequest(ctx, "/v1/completions", req)
	if err != nil {
		return "", fmt.Errorf("failed to generate text: %w", err)
	}

	var generateResp GenerateResponse
	if err := json.Unmarshal(resp, &generateResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(generateResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return generateResp.Choices[0].Text, nil
}

// GenerateWithTools generates text and can use tools
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	// For now, vLLM doesn't support tool calling in the same way as OpenAI/Anthropic
	// We'll implement a basic version that includes tool descriptions in the prompt
	if len(tools) == 0 {
		return c.Generate(ctx, prompt, options...)
	}

	// Build tool descriptions
	var toolDescriptions []string
	for _, tool := range tools {
		toolDescriptions = append(toolDescriptions, fmt.Sprintf("- %s: %s", tool.Name(), tool.Description()))
	}

	// Create enhanced prompt with tool information
	enhancedPrompt := fmt.Sprintf(`%s

Available tools:
%s

Please respond to the user's request. If you need to use any tools, describe what you would do.`, prompt, strings.Join(toolDescriptions, "\n"))

	return c.Generate(ctx, enhancedPrompt, options...)
}

// GenerateDetailed generates text and returns detailed response information including token usage
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	// Call the existing method and construct a detailed response
	content, err := c.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}

	// Return a detailed response without usage information (vLLM doesn't provide token usage)
	return &contracts.LLMResponse{
		Content:    content,
		Model:      c.Model,
		StopReason: "",
		Usage:      nil, // vLLM doesn't provide token usage information
		Metadata: map[string]interface{}{
			"provider": "vllm",
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
		Usage:      nil, // vLLM doesn't provide token usage information
		Metadata: map[string]interface{}{
			"provider":   "vllm",
			"tools_used": true,
		},
	}, nil
}

package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// GenerateWithTools implements contracts.LLM.GenerateWithTools
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	if c.Model == "" {
		return "", fmt.Errorf("model not specified: use WithModel option when creating the client")
	}
	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, opt := range options {
		if opt != nil {
			opt(params)
		}
	}
	maxIterations := params.MaxIterations
	if maxIterations == 0 {
		maxIterations = 2
	}
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, "default")
	}

	anthropicTools := anthropicConvertTools(tools)
	toolCallHistory := make(map[string]int)
	messages := c.buildMessagesWithMemory(ctx, prompt, params)
	maxTokens := 2048
	if params.LLMConfig != nil && params.LLMConfig.EnableReasoning && params.LLMConfig.ReasoningBudget > 0 {
		maxTokens = params.LLMConfig.ReasoningBudget + 4000
	}
	var lastContent string

	for iteration := 0; iteration < maxIterations; iteration++ {
		req := CompletionRequest{
			Model: c.Model, Messages: messages, MaxTokens: maxTokens,
			Temperature: params.LLMConfig.Temperature, TopP: params.LLMConfig.TopP,
			Tools: anthropicTools, ToolChoice: map[string]string{"type": "auto"},
		}
		if params.SystemMessage != "" {
			if params.ResponseFormat != nil {
				req.System = params.SystemMessage + "\n\nIMPORTANT: You must respond with valid JSON that matches the specified schema. Return ONLY the raw JSON object without any markdown formatting, code blocks, or wrapper text. Pay special attention to array fields - if a field is defined as an array in the schema, it MUST be an array in your response, not an object."
			} else {
				req.System = params.SystemMessage
			}
			c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
		} else if params.ResponseFormat != nil {
			req.System = "You must respond with valid JSON that matches the specified schema. Return ONLY the raw JSON object without any markdown formatting, code blocks, or wrapper text. Pay special attention to array fields - if a field is defined as an array in the schema, it MUST be an array in your response, not an object."
			c.logger.Debug(ctx, "Added system message for structured output", nil)
		}
		if params.LLMConfig != nil && params.LLMConfig.Reasoning != "" {
			c.logger.Debug(ctx, "Reasoning mode not supported in current API version", map[string]interface{}{"reasoning": params.LLMConfig.Reasoning})
		}
		c.logger.Debug(ctx, "Sending request with tools to Anthropic", map[string]interface{}{
			"model": c.Model, "temperature": req.Temperature, "top_p": req.TopP,
			"messages": len(req.Messages), "tools": len(req.Tools), "system": req.System != "",
			"iteration": iteration + 1, "maxIterations": maxIterations,
		})

		resp, err := c.invokeToolRequest(ctx, &req, params, iteration)
		if err != nil {
			return "", err
		}
		if resp.Content == nil {
			c.logger.Error(ctx, "No content in response", map[string]interface{}{"iteration": iteration + 1})
			return "", fmt.Errorf("no content in response (iteration %d)", iteration+1)
		}

		hasToolUse, toolCalls, textContent := c.inspectToolResponse(ctx, resp, iteration)
		if len(textContent) > 0 {
			lastContent = strings.Join(textContent, "\n")
		}
		if !hasToolUse {
			if len(textContent) == 0 {
				return "", fmt.Errorf("no text content in response (iteration %d)", iteration+1)
			}
			response := strings.Join(textContent, "\n")
			if params.ResponseFormat != nil {
				extractedJSON := anthropicExtractJSONFromResponse(response)
				if extractedJSON != response {
					c.logger.Debug(ctx, "Extracted JSON from response", map[string]interface{}{
						"original_length": len(response), "extracted_length": len(extractedJSON),
					})
					response = extractedJSON
				}
			}
			c.logger.Debug(ctx, "Returning final response (no tool use)", map[string]interface{}{
				"response_length": len(response), "response_preview": response, "iteration": iteration + 1,
			})
			return response, nil
		}

		c.logger.Info(ctx, "Processing tool calls", map[string]interface{}{"count": len(toolCalls), "iteration": iteration + 1})
		assistantContent := strings.Join(textContent, "\n")
		if strings.TrimSpace(assistantContent) != "" {
			messages = append(messages, Message{Role: "assistant", Content: assistantContent})
		}
		toolResults := c.executeToolsParallel(ctx, toolCalls, tools, params, toolCallHistory, iteration)
		toolResultsJSON, err := json.Marshal(toolResults)
		if err != nil {
			return "", fmt.Errorf("failed to marshal tool results (iteration %d): %w", iteration+1, err)
		}
		messages = append(messages, Message{Role: "user", Content: fmt.Sprintf("Here are the tool results: %s", string(toolResultsJSON))})
	}

	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final summary call", map[string]interface{}{"maxIterations": maxIterations})
		return lastContent, nil
	}
	return c.generateFinalToolResponse(ctx, messages, params, maxTokens, maxIterations)
}

func anthropicConvertTools(tools []contracts.Tool) []Tool {
	anthropicTools := make([]Tool, len(tools))
	for i, tool := range tools {
		properties := make(map[string]interface{})
		required := []string{}
		for name, param := range tool.Parameters() {
			properties[name] = map[string]interface{}{"type": param.Type, "description": param.Description}
			if param.Default != nil {
				properties[name].(map[string]interface{})["default"] = param.Default
			}
			if param.Required {
				required = append(required, name)
			}
			if param.Items != nil {
				properties[name].(map[string]interface{})["items"] = map[string]interface{}{"type": param.Items.Type}
				if param.Items.Enum != nil {
					properties[name].(map[string]interface{})["items"].(map[string]interface{})["enum"] = param.Items.Enum
				}
			}
			if param.Enum != nil {
				properties[name].(map[string]interface{})["enum"] = param.Enum
			}
		}
		anthropicTools[i] = Tool{
			Name: tool.Name(), Description: tool.Description(),
			InputSchema: map[string]interface{}{"type": "object", "properties": properties, "required": required},
		}
	}
	return anthropicTools
}

// GenerateWithToolsDetailed generates text with tools and returns detailed response information including token usage
func (c *Client) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	content, err := c.GenerateWithTools(ctx, prompt, tools, options...)
	if err != nil {
		return nil, err
	}
	return &contracts.LLMResponse{
		Content: content, Model: c.Model, StopReason: "", Usage: nil,
		Metadata: map[string]interface{}{"provider": "anthropic", "tools_used": true},
	}, nil
}

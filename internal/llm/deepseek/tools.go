package deepseek

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// GenerateWithTools implements contracts.LLM.GenerateWithTools
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	response, err := c.GenerateWithToolsDetailed(ctx, prompt, tools, options...)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// GenerateWithToolsDetailed implements contracts.LLM.GenerateWithToolsDetailed
func (c *Client) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	// Apply options
	params := &contracts.GenerateOptions{
		LLMConfig: &contracts.LLMConfig{
			Temperature: 0.7,
		},
	}

	for _, option := range options {
		option(params)
	}

	// Set default max iterations if not provided
	maxIterations := params.MaxIterations
	if maxIterations == 0 {
		maxIterations = DefaultMaxIterations
	}

	// Get organization ID from context if available
	orgID, _ := multitenancy.GetOrgID(ctx)

	// Convert tools to DeepSeek format
	deepseekTools := c.convertToolsToDeepSeekFormat(tools)

	// Build initial messages
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

	// Track total token usage across all iterations
	var totalInputTokens, totalOutputTokens int

	// Iterative tool calling loop
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Create request
		req := ChatCompletionRequest{
			Model:    c.Model,
			Messages: messages,
			Tools:    deepseekTools,
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

		// Set response format if provided (only on last iteration)
		if params.ResponseFormat != nil && iteration == maxIterations-1 {
			req.ResponseFormat = &ResponseFormatParam{
				Type:       "json_schema",
				JSONSchema: params.ResponseFormat.Schema,
			}
		}

		c.logger.Debug(ctx, "Sending request with tools to DeepSeek", map[string]interface{}{
			"model":         c.Model,
			"temperature":   req.Temperature,
			"messages":      len(req.Messages),
			"tools":         len(req.Tools),
			"iteration":     iteration + 1,
			"maxIterations": maxIterations,
			"org_id":        orgID,
		})

		// Make request
		resp, err := c.doRequest(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from DeepSeek API", map[string]interface{}{
				"error": err.Error(),
				"model": c.Model,
			})
			return nil, fmt.Errorf("failed to generate text with tools: %w", err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("no response from DeepSeek API")
		}

		// Accumulate token usage
		totalInputTokens += resp.Usage.PromptTokens
		totalOutputTokens += resp.Usage.CompletionTokens

		// Check if the model wants to use tools
		if len(resp.Choices[0].Message.ToolCalls) == 0 {
			// No tool calls, return the final response
			c.logger.Debug(ctx, "No tool calls, returning final response", map[string]interface{}{
				"iteration": iteration + 1,
			})

			return &contracts.LLMResponse{
				Content:    resp.Choices[0].Message.Content,
				Model:      resp.Model,
				StopReason: resp.Choices[0].FinishReason,
				Usage: &contracts.TokenUsage{
					InputTokens:  totalInputTokens,
					OutputTokens: totalOutputTokens,
					TotalTokens:  totalInputTokens + totalOutputTokens,
				},
				Metadata: map[string]interface{}{
					"provider":   "deepseek",
					"iterations": iteration + 1,
				},
			}, nil
		}

		// The model wants to use tools
		toolCalls := resp.Choices[0].Message.ToolCalls
		c.logger.Info(ctx, "Processing tool calls", map[string]interface{}{
			"count":     len(toolCalls),
			"iteration": iteration + 1,
		})

		// Add the assistant's message with tool calls to the conversation
		messages = append(messages, resp.Choices[0].Message)

		// Store assistant message with tool calls in memory
		if params.Memory != nil {
			memToolCalls := make([]contracts.ToolCall, len(toolCalls))
			for i, tc := range toolCalls {
				memToolCalls[i] = contracts.ToolCall{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				}
			}
			_ = params.Memory.AddMessage(ctx, contracts.Message{
				Role:      contracts.RoleAssistant,
				Content:   resp.Choices[0].Message.Content,
				ToolCalls: memToolCalls,
			})
		}

		// Execute tools in parallel
		toolResults := c.executeToolsParallel(ctx, toolCalls, tools)

		// Add tool results to messages and memory
		for _, result := range toolResults {
			messages = append(messages, Message{
				Role:       "tool",
				Content:    result.Content,
				ToolCallID: result.ToolCallID,
				Name:       result.ToolName,
			})

			// Store tool result in memory
			if params.Memory != nil {
				_ = params.Memory.AddMessage(ctx, contracts.Message{
					Role:       contracts.MessageRoleTool,
					Content:    result.Content,
					ToolCallID: result.ToolCallID,
					Metadata: map[string]interface{}{
						"tool_name": result.ToolName,
					},
				})
			}
		}
	}

	// If we've exhausted max iterations, make one final call without tools
	c.logger.Warn(ctx, "Max iterations reached, making final call without tools", map[string]interface{}{
		"maxIterations": maxIterations,
	})

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

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to make final request: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from DeepSeek API")
	}

	totalInputTokens += resp.Usage.PromptTokens
	totalOutputTokens += resp.Usage.CompletionTokens

	return &contracts.LLMResponse{
		Content:    resp.Choices[0].Message.Content,
		Model:      resp.Model,
		StopReason: resp.Choices[0].FinishReason,
		Usage: &contracts.TokenUsage{
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
			TotalTokens:  totalInputTokens + totalOutputTokens,
		},
		Metadata: map[string]interface{}{
			"provider":       "deepseek",
			"iterations":     maxIterations,
			"max_iterations": true,
		},
	}, nil
}

package openai

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// GenerateWithTools implements contracts.LLM.GenerateWithTools.
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (string, error) {
	params := &contracts.GenerateOptions{}
	for _, opt := range options {
		if opt != nil {
			opt(params)
		}
	}
	if params.LLMConfig == nil {
		params.LLMConfig = &contracts.LLMConfig{
			Temperature: 0.7, TopP: 1.0, FrequencyPenalty: 0.0, PresencePenalty: 0.0,
		}
	}
	maxIterations := params.MaxIterations
	if maxIterations == 0 {
		maxIterations = 2
	}

	orgID := "default"
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		orgID = id
	}
	ctx = context.WithValue(ctx, openAIOrganizationKey, orgID)

	messages, req := c.newToolRequest(ctx, prompt, tools, params)
	toolCallHistory := make(map[string]int)
	var toolCallHistoryMu sync.Mutex
	var lastContent string

	for iteration := 0; iteration < maxIterations; iteration++ {
		req.Messages = messages
		reasoningEffort := "none"
		if params.LLMConfig != nil && params.LLMConfig.Reasoning != "" {
			reasoningEffort = params.LLMConfig.Reasoning
		}
		c.logger.Debug(ctx, "Sending request with tools to OpenAI", map[string]interface{}{
			"model": c.Model, "temperature": req.Temperature, "top_p": req.TopP,
			"frequency_penalty": req.FrequencyPenalty, "presence_penalty": req.PresencePenalty,
			"stop_sequences": req.Stop, "messages": len(req.Messages), "tools": len(req.Tools),
			"response_format": params.ResponseFormat != nil, "parallel_tools": req.ParallelToolCalls,
			"reasoning_effort": reasoningEffort, "iteration": iteration + 1, "maxIterations": maxIterations,
		})
		resp, err := c.ChatService.Completions.New(ctx, req)
		if err != nil {
			c.logger.Error(ctx, "Error from OpenAI API", map[string]interface{}{"error": err.Error()})
			return "", fmt.Errorf("failed to create chat completion: %w", err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no completions returned")
		}
		if acc := openAIGetUsageAccumulator(ctx); acc != nil {
			acc.add(int(resp.Usage.PromptTokens), int(resp.Usage.CompletionTokens), int(resp.Usage.TotalTokens),
				int(resp.Usage.CompletionTokensDetails.ReasoningTokens), c.Model)
		}

		lastContent = strings.TrimSpace(resp.Choices[0].Message.Content)
		if len(resp.Choices[0].Message.ToolCalls) == 0 {
			return lastContent, nil
		}
		toolCalls := resp.Choices[0].Message.ToolCalls
		c.logger.Info(ctx, "Processing tool calls", map[string]interface{}{"count": len(toolCalls), "iteration": iteration + 1})
		messages = append(messages, resp.Choices[0].Message.ToParam())

		for _, toolCall := range toolCalls {
			if toolCall.Function.Name == "multi_tool_use.parallel" {
				c.logger.Info(ctx, "Replacing multi_tool_use.parallel with parallel_tool_use", nil)
				toolCall.Function.Name = "parallel_tool_use"
			}
			if toolCall.Function.Name == "parallel_tool_use" {
				message, ok, err := c.executeParallelToolCall(ctx, toolCall, tools, params.Memory, toolCallHistory, &toolCallHistoryMu)
				if err != nil {
					return "", err
				}
				if ok {
					messages = append(messages, message)
				}
				continue
			}
			messages = append(messages, c.executeToolCall(ctx, toolCall, tools, params.Memory, toolCallHistory, &toolCallHistoryMu, resp))
		}
	}

	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final summary call", map[string]interface{}{"maxIterations": maxIterations})
		return lastContent, nil
	}
	return c.finalToolResponse(ctx, messages, params, maxIterations)
}

// GenerateWithToolsDetailed returns tool-loop token usage aggregated across calls.
func (c *Client) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	acc := &openAIUsageAccumulator{}
	ctx = openAIWithUsageAccumulator(ctx, acc)
	content, err := c.GenerateWithTools(ctx, prompt, tools, options...)
	if err != nil {
		return nil, err
	}
	usage, model, _ := acc.snapshot()
	if model == "" {
		model = c.Model
	}
	return &contracts.LLMResponse{
		Content: content, Model: model, StopReason: "", Usage: usage,
		Metadata: map[string]interface{}{"provider": "openai", "tools_used": true},
	}, nil
}

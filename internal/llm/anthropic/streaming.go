package anthropic

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// GenerateStream implements contracts.StreamingLLM.GenerateStream
func (c *Client) GenerateStream(ctx context.Context, prompt string, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	c.logger.Debug(ctx, "[LLM RESPONSE DEBUG] GenerateStream called (NO TOOLS)", map[string]interface{}{
		"model": c.Model, "promptLength": len(prompt),
	})
	if c.Model == "" {
		return nil, fmt.Errorf("model not specified: use WithModel option when creating the client")
	}

	params := &contracts.GenerateOptions{LLMConfig: &contracts.LLMConfig{Temperature: 0.7}}
	for _, option := range options {
		option(params)
	}

	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, "default")
	}

	builder := anthropicNewMessageHistoryBuilder(c.logger)
	messages := builder.buildMessages(ctx, prompt, params)
	maxTokens := 2048
	if params.LLMConfig != nil && params.LLMConfig.EnableReasoning && params.LLMConfig.ReasoningBudget > 0 {
		maxTokens = params.LLMConfig.ReasoningBudget + 4000
	}

	req := CompletionRequest{
		Model: c.Model, Messages: messages, MaxTokens: maxTokens,
		Temperature: params.LLMConfig.Temperature, TopP: params.LLMConfig.TopP, Stream: true,
	}
	if params.SystemMessage != "" {
		req.System = params.SystemMessage
		c.logger.Debug(ctx, "Using system message", map[string]interface{}{"system_message": params.SystemMessage})
	}
	if params.LLMConfig != nil && len(params.LLMConfig.StopSequences) > 0 {
		req.StopSequences = params.LLMConfig.StopSequences
	}
	if params.LLMConfig != nil && params.LLMConfig.EnableReasoning {
		if SupportsThinking(c.Model) {
			req.Thinking = &ReasoningSpec{Type: "enabled"}
			if params.LLMConfig.ReasoningBudget > 0 {
				req.Thinking.BudgetTokens = params.LLMConfig.ReasoningBudget
			}
			req.Temperature = 1.0
			c.logger.Debug(ctx, "Enabled reasoning (thinking) tokens", map[string]interface{}{
				"model": c.Model, "budget_tokens": params.LLMConfig.ReasoningBudget,
				"max_tokens": maxTokens, "temperature": req.Temperature,
			})
		} else {
			c.logger.Warn(ctx, "Thinking tokens not supported by this model", map[string]interface{}{
				"model":            c.Model,
				"supported_models": []string{"claude-3-7-sonnet-20250219", "claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-opus-4-1-20250805"},
			})
		}
	}

	bufferSize := 100
	if params.StreamConfig != nil {
		bufferSize = params.StreamConfig.BufferSize
	}
	eventChan := make(chan contracts.StreamEvent, bufferSize)
	go func() {
		defer func() {
			defer func() { _ = recover() }()
			close(eventChan)
		}()

		c.logger.Debug(ctx, "[LLM RESPONSE DEBUG] Executing streaming request without tools", map[string]interface{}{
			"model": c.Model, "hasMemory": params != nil && params.Memory != nil, "temperature": req.Temperature,
		})
		if err := c.executeStreamingRequestWithMemory(ctx, req, eventChan, prompt, params); err != nil {
			c.logger.Error(ctx, "[LLM RESPONSE DEBUG] Streaming request failed", map[string]interface{}{"error": err.Error()})
			select {
			case eventChan <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: err, Timestamp: time.Now()}:
			case <-ctx.Done():
				return
			}
		} else {
			c.logger.Info(ctx, "[LLM RESPONSE DEBUG] Streaming request completed successfully (no tools)", map[string]interface{}{"model": c.Model})
		}
	}()
	return eventChan, nil
}

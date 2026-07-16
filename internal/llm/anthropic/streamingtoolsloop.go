package anthropic

import (
	"context"
	"strings"
	"time"

	"nu/internal/contracts"
)

// executeStreamingWithTools handles streaming requests with tools using iterative loop
func (c *Client) executeStreamingWithTools(ctx context.Context, prompt string, anthropicTools []Tool, originalTools []contracts.Tool, params *contracts.GenerateOptions, eventChan chan<- contracts.StreamEvent) error {
	builder := anthropicNewMessageHistoryBuilder(c.logger)
	messages := builder.buildMessages(ctx, prompt, params)
	maxIterations := 2
	if params.MaxIterations > 0 {
		maxIterations = params.MaxIterations
	}
	maxTokens := 2048
	if params.LLMConfig != nil && params.LLMConfig.EnableReasoning && params.LLMConfig.ReasoningBudget > 0 {
		maxTokens = params.LLMConfig.ReasoningBudget + 4000
	}

	gotCompleteResponse := false
	finalIterationCount := 0
	for iteration := 0; iteration < maxIterations; iteration++ {
		finalIterationCount = iteration + 1
		req := CompletionRequest{
			Model: c.Model, Messages: messages, MaxTokens: maxTokens,
			Temperature: params.LLMConfig.Temperature, TopP: params.LLMConfig.TopP,
			Tools: anthropicTools, ToolChoice: map[string]string{"type": "auto"}, Stream: true,
		}
		if params.SystemMessage != "" {
			req.System = params.SystemMessage
		}
		if params.LLMConfig != nil && params.LLMConfig.EnableReasoning && SupportsThinking(c.Model) {
			req.Thinking = &ReasoningSpec{Type: "enabled"}
			if params.LLMConfig.ReasoningBudget > 0 {
				req.Thinking.BudgetTokens = params.LLMConfig.ReasoningBudget
			}
			req.Temperature = 1.0
			c.logger.Debug(ctx, "Enabled reasoning (thinking) tokens for tools", map[string]interface{}{
				"model": c.Model, "budget_tokens": params.LLMConfig.ReasoningBudget,
				"max_tokens": maxTokens, "temperature": req.Temperature,
				"iteration": iteration + 1, "maxIterations": maxIterations,
			})
		}

		c.logger.Debug(ctx, "[LLM RESPONSE DEBUG] Calling LLM for iteration", map[string]interface{}{
			"iteration": iteration + 1, "maxIterations": maxIterations, "hasTools": len(anthropicTools) > 0,
		})
		filterContentDeltas := true
		if params.StreamConfig != nil && params.StreamConfig.IncludeIntermediateMessages {
			filterContentDeltas = false
		}
		toolCalls, hasContent, capturedContentEvents, err := c.executeStreamingRequestWithToolCapture(ctx, req, eventChan, filterContentDeltas, params)
		if err != nil {
			c.logger.Error(ctx, "[LLM RESPONSE DEBUG] LLM call failed", map[string]interface{}{"iteration": iteration + 1, "error": err.Error()})
			return err
		}
		c.logger.Info(ctx, "[LLM RESPONSE DEBUG] LLM response received", map[string]interface{}{
			"iteration": iteration + 1, "toolCallsCount": len(toolCalls), "hasContent": hasContent, "gotToolCalls": len(toolCalls) > 0,
		})

		if len(toolCalls) == 0 {
			if hasContent {
				c.logger.Info(ctx, "[LLM RESPONSE DEBUG] Got final content response without tool calls", map[string]interface{}{
					"iteration": iteration + 1, "hasContent": hasContent, "responseType": "final_answer",
					"toolCallsCount": 0, "capturedEvents": len(capturedContentEvents),
				})
				if filterContentDeltas {
					c.logger.Debug(ctx, "[LLM RESPONSE DEBUG] Replaying captured content events", map[string]interface{}{
						"iteration": iteration + 1, "eventsCount": len(capturedContentEvents),
					})
					for _, contentEvent := range capturedContentEvents {
						select {
						case eventChan <- contentEvent:
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}
				select {
				case eventChan <- contracts.StreamEvent{Type: contracts.StreamEventContentComplete, Timestamp: time.Now(), Metadata: map[string]interface{}{"iteration": iteration + 1}}:
				case <-ctx.Done():
					return ctx.Err()
				}
				gotCompleteResponse = true
				break
			}
			c.logger.Warn(ctx, "[LLM RESPONSE DEBUG] No tool calls and no content in iteration", map[string]interface{}{
				"iteration": iteration + 1, "maxIterations": maxIterations, "responseType": "empty_response",
			})
			if iteration >= maxIterations-1 {
				break
			}
			continue
		}

		c.logger.Info(ctx, "[LLM RESPONSE DEBUG] Processing tool calls from LLM response", map[string]interface{}{
			"count": len(toolCalls), "iteration": iteration + 1, "responseType": "tool_calls",
		})
		var assistantContent strings.Builder
		for _, event := range capturedContentEvents {
			if event.Type == contracts.StreamEventContentDelta {
				assistantContent.WriteString(event.Content)
			}
		}
		if strings.TrimSpace(assistantContent.String()) != "" {
			messages = append(messages, Message{Role: "assistant", Content: assistantContent.String()})
		}

		if err := c.executeStreamingTools(ctx, toolCalls, originalTools, &messages, eventChan, iteration); err != nil {
			return err
		}
		if iteration < maxIterations-1 {
			select {
			case eventChan <- contracts.StreamEvent{
				Type: contracts.StreamEventContentDelta, Content: "\n\n", Timestamp: time.Now(),
				Metadata: map[string]interface{}{"iteration_boundary": true, "iteration": iteration + 1},
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	if gotCompleteResponse {
		c.logger.Debug(ctx, "[LLM RESPONSE DEBUG] Skipping final synthesis call - already got complete response", map[string]interface{}{
			"maxIterations": maxIterations, "totalLLMCalls": finalIterationCount, "skippedSynthesisCall": true,
		})
		return nil
	}
	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final synthesis call", map[string]interface{}{"maxIterations": maxIterations})
		return nil
	}
	return c.executeFinalToolStream(ctx, messages, params, eventChan, maxTokens, maxIterations, finalIterationCount)
}

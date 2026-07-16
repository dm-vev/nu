package tools

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// Run executes the tool with the given input
func (at *AgentTool) Run(ctx context.Context, input string) (string, error) {
	startTime := time.Now()
	agentName := at.agent.GetName()

	// Start tracing span if tracer is available
	var span contracts.Span
	if at.tracer != nil {
		ctx, span = at.tracer.StartSpan(ctx, fmt.Sprintf("sub_agent.%s", agentName))
		defer span.End()

		// Add span attributes
		span.SetAttribute("sub_agent.name", agentName)
		span.SetAttribute("sub_agent.input", input)
		span.SetAttribute("sub_agent.tool_name", at.name)
	}

	// Add agent name to context for tracing
	ctx = telemetry.WithAgentName(ctx, agentName)

	// Check recursion depth
	depth := getRecursionDepth(ctx)
	if depth > MaxRecursionDepth {
		err := fmt.Errorf("maximum recursion depth %d exceeded (current: %d)", MaxRecursionDepth, depth)
		if span != nil {
			span.AddEvent("error", map[string]interface{}{
				"error": err.Error(),
			})
			span.SetAttribute("sub_agent.error", err.Error())
		}
		at.logger.Error(ctx, "Sub-agent recursion depth exceeded", map[string]interface{}{
			"sub_agent":       agentName,
			"recursion_depth": depth,
			"max_depth":       MaxRecursionDepth,
		})
		return "", err
	}

	// Update context with sub-agent metadata
	ctx = context.WithValue(ctx, subAgentNameKey, agentName)
	ctx = context.WithValue(ctx, parentAgentKey, "main")
	ctx = context.WithValue(ctx, recursionDepthKey, depth+1)

	// Check if parent context has a deadline that would expire before our timeout
	var cancel context.CancelFunc
	parentDeadline, hasDeadline := ctx.Deadline()
	desiredDeadline := time.Now().Add(at.timeout)

	if hasDeadline && parentDeadline.Before(desiredDeadline) {
		// Parent context has a shorter deadline - we need to extend it
		// Create a new context that preserves values but has our longer timeout
		at.logger.Warn(ctx, "Parent context has shorter deadline, extending timeout for sub-agent", map[string]interface{}{
			"parent_deadline": parentDeadline.Format(time.RFC3339),
			"desired_timeout": at.timeout.String(),
			"sub_agent":       agentName,
		})

		// Use context.WithoutCancel to remove parent's deadline while preserving values
		// This is available in Go 1.21+, otherwise we need to manually copy values
		newCtx := context.WithoutCancel(ctx)
		ctx, cancel = context.WithTimeout(newCtx, at.timeout)
	} else {
		// Parent context doesn't have a shorter deadline, use normal timeout
		ctx, cancel = context.WithTimeout(ctx, at.timeout)
	}
	defer cancel()

	// Log sub-agent invocation with debug details
	at.logger.Debug(ctx, "Invoking sub-agent", map[string]interface{}{
		"sub_agent":       agentName,
		"tool_name":       at.name,
		"input_prompt":    input,
		"recursion_depth": depth + 1,
		"timeout":         at.timeout.String(),
	})

	// Check if we have a stream forwarder in the context
	var response *contracts.AgentResponse
	var err error

	if forwarder, ok := ctx.Value(contracts.StreamForwarderKey).(contracts.StreamForwarder); ok && forwarder != nil {
		// Use streaming to forward events to parent
		result, streamErr := at.runWithStreaming(ctx, input, forwarder, span, agentName)
		if streamErr != nil {
			err = streamErr
		} else {
			// After streaming completes, get detailed response for tracking
			response, err = at.agent.RunDetailed(ctx, input)
			if err == nil && response.Content == "" {
				// If detailed response is empty, use streamed result
				response.Content = result
			}
		}
	} else {
		// Fall back to detailed execution for full tracking
		response, err = at.agent.RunDetailed(ctx, input)
	}

	duration := time.Since(startTime)

	if err != nil {
		// Log error details
		at.logger.Error(ctx, "Sub-agent execution failed", map[string]interface{}{
			"sub_agent": agentName,
			"tool_name": at.name,
			"error":     err.Error(),
			"duration":  duration.String(),
			"input":     input,
		})

		// Record error in span
		if span != nil {
			span.AddEvent("error", map[string]interface{}{
				"error": err.Error(),
			})
			span.SetAttribute("sub_agent.error", err.Error())
			span.SetAttribute("sub_agent.duration_ms", duration.Milliseconds())
		}

		return "", fmt.Errorf("sub-agent %s failed: %w", agentName, err)
	}

	// Log comprehensive execution details
	executionDetails := map[string]interface{}{
		"sub_agent":         agentName,
		"tool_name":         at.name,
		"input_prompt":      input,
		"response_content":  response.Content,
		"response_length":   len(response.Content),
		"duration":          duration.String(),
		"agent_name":        response.AgentName,
		"model_used":        response.Model,
		"llm_calls":         response.ExecutionSummary.LLMCalls,
		"tool_calls":        response.ExecutionSummary.ToolCalls,
		"sub_agent_calls":   response.ExecutionSummary.SubAgentCalls,
		"execution_time_ms": response.ExecutionSummary.ExecutionTimeMs,
		"used_tools":        response.ExecutionSummary.UsedTools,
		"used_sub_agents":   response.ExecutionSummary.UsedSubAgents,
	}

	// Add token usage details if available
	if response.Usage != nil {
		executionDetails["input_tokens"] = response.Usage.InputTokens
		executionDetails["output_tokens"] = response.Usage.OutputTokens
		executionDetails["total_tokens"] = response.Usage.TotalTokens
		executionDetails["reasoning_tokens"] = response.Usage.ReasoningTokens
	}

	// Add metadata if available
	if response.Metadata != nil {
		executionDetails["metadata"] = response.Metadata
	}

	at.logger.Info(ctx, "Sub-agent execution completed with detailed tracking", executionDetails)

	// Record detailed success information in span
	if span != nil {
		span.SetAttribute("sub_agent.response", response.Content)
		span.SetAttribute("sub_agent.duration_ms", duration.Milliseconds())
		span.SetAttribute("sub_agent.response_length", len(response.Content))
		span.SetAttribute("sub_agent.success", true)
		span.SetAttribute("sub_agent.agent_name", response.AgentName)
		span.SetAttribute("sub_agent.model_used", response.Model)
		span.SetAttribute("sub_agent.llm_calls", response.ExecutionSummary.LLMCalls)
		span.SetAttribute("sub_agent.tool_calls", response.ExecutionSummary.ToolCalls)
		span.SetAttribute("sub_agent.sub_agent_calls", response.ExecutionSummary.SubAgentCalls)
		span.SetAttribute("sub_agent.execution_time_ms", response.ExecutionSummary.ExecutionTimeMs)

		// Add token usage to span if available
		if response.Usage != nil {
			span.SetAttribute("sub_agent.input_tokens", response.Usage.InputTokens)
			span.SetAttribute("sub_agent.output_tokens", response.Usage.OutputTokens)
			span.SetAttribute("sub_agent.total_tokens", response.Usage.TotalTokens)
			span.SetAttribute("sub_agent.reasoning_tokens", response.Usage.ReasoningTokens)
		}
	}

	return response.Content, nil
}

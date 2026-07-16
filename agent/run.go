package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dm-vev/nu/agent/execution"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
)

// Run runs the agent with the given input.
func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	response, err := a.runInternal(ctx, input, false)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// RunDetailed runs the agent and returns usage and execution metadata.
func (a *Agent) RunDetailed(ctx context.Context, input string) (*contracts.AgentResponse, error) {
	return a.runInternal(ctx, input, true)
}

func (a *Agent) runInternal(ctx context.Context, input string, detailed bool) (*contracts.AgentResponse, error) {
	startTime := time.Now()
	tracker := execution.NewTracker(detailed)
	ctx = execution.WithTracker(ctx, tracker)

	var (
		response string
		err      error
	)
	switch {
	case a.customRunFunc != nil:
		response, err = a.customRunFunc(ctx, input, a)
	case a.isRemote:
		response, err = a.runRemoteWithTracking(ctx, input)
	default:
		response, err = a.runLocalWithTracking(ctx, input)
	}
	if err != nil {
		return nil, err
	}

	tracker.SetExecutionTime(time.Since(startTime).Milliseconds())
	usage, execSummary, primaryModel := tracker.Results()
	var summary contracts.ExecutionSummary
	if execSummary != nil {
		summary = *execSummary
	}
	if detailed {
		logExecutionDetails(a, input, response, primaryModel, startTime, usage, execSummary)
	}
	return &contracts.AgentResponse{
		Content:          response,
		Usage:            usage,
		AgentName:        a.name,
		Model:            primaryModel,
		ExecutionSummary: summary,
		Metadata: map[string]interface{}{
			"agent_name":            a.name,
			"execution_timestamp":   startTime.Unix(),
			"execution_duration_ms": time.Since(startTime).Milliseconds(),
		},
	}, nil
}

func logExecutionDetails(a *Agent, input, response, model string, startTime time.Time, usage *contracts.TokenUsage, summary *contracts.ExecutionSummary) {
	details := map[string]interface{}{
		"agent_name":        a.name,
		"input_length":      len(input),
		"response_length":   len(response),
		"model_used":        model,
		"execution_time_ms": time.Since(startTime).Milliseconds(),
	}
	if summary != nil {
		details["llm_calls"] = summary.LLMCalls
		details["tool_calls"] = summary.ToolCalls
		details["sub_agent_calls"] = summary.SubAgentCalls
		details["used_tools"] = summary.UsedTools
		details["used_sub_agents"] = summary.UsedSubAgents
	}
	if usage != nil {
		details["input_tokens"] = usage.InputTokens
		details["output_tokens"] = usage.OutputTokens
		details["total_tokens"] = usage.TotalTokens
		details["reasoning_tokens"] = usage.ReasoningTokens
	}
	log.Printf("[Agent SDK] Agent execution completed: %+v", details)
}

func (a *Agent) runLocalWithTracking(ctx context.Context, input string) (string, error) {
	ctx = telemetry.WithAgentName(ctx, a.name)
	if a.orgID != "" {
		ctx = multitenancy.WithOrgID(ctx, a.orgID)
	}

	var span contracts.Span
	if a.tracer != nil {
		ctx, span = a.tracer.StartSpan(ctx, "agent.Run")
		defer span.End()
	}
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{Role: contracts.RoleUser, Content: input}); err != nil {
			return "", fmt.Errorf("failed to add user message to memory: %w", err)
		}
	}
	if a.guardrails != nil {
		guardedInput, err := a.guardrails.ProcessInput(ctx, input)
		if err != nil {
			return "", fmt.Errorf("guardrails error: %w", err)
		}
		input = guardedInput
	}

	taskID, action, planInput := a.extractPlanAction(input)
	if taskID != "" {
		return a.handlePlanAction(ctx, taskID, action, planInput)
	}
	if a.systemPrompt != "" && a.isAskingAboutRole(input) {
		response := a.generateRoleResponse()
		if a.memory != nil {
			if err := a.memory.AddMessage(ctx, contracts.Message{Role: contracts.RoleAssistant, Content: response}); err != nil {
				return "", fmt.Errorf("failed to add role response to memory: %w", err)
			}
		}
		return response, nil
	}

	allTools := a.tools
	if len(a.mcpServers) > 0 {
		mcpTools, err := a.collectMCPTools(ctx)
		if err != nil {
			a.logger.Warn(context.Background(), fmt.Sprintf("Failed to collect MCP tools: %v", err), nil)
		} else if len(mcpTools) > 0 {
			allTools = deduplicateTools(append(allTools, mcpTools...))
		}
	}
	if len(a.lazyMCPConfigs) > 0 {
		allTools = deduplicateTools(append(allTools, a.createLazyMCPTools()...))
	}
	if len(allTools) > 0 && a.requirePlanApproval {
		a.planService.ResetTools(a.llm, allTools, a.systemPrompt, a.requirePlanApproval)
		return a.runWithExecutionPlan(ctx, input)
	}
	return a.runWithoutExecutionPlanWithToolsTracked(ctx, input, allTools)
}

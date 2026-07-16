package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"nu/internal/agent/plans"
	"nu/internal/contracts"
	"nu/internal/multitenancy"
	"nu/internal/telemetry"
)

// Run runs the agent with the given input
func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	response, err := a.runInternal(ctx, input, false)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

func (a *Agent) RunDetailed(ctx context.Context, input string) (*contracts.AgentResponse, error) {
	return a.runInternal(ctx, input, true)
}

func (a *Agent) runInternal(ctx context.Context, input string, detailed bool) (*contracts.AgentResponse, error) {
	startTime := time.Now()

	tracker := newUsageTracker(detailed)
	ctx = withUsageTracker(ctx, tracker)

	var response string
	var err error

	if a.customRunFunc != nil {
		response, err = a.customRunFunc(ctx, input, a)
		if err != nil {
			return nil, err
		}
	} else if a.isRemote {
		response, err = a.runRemoteWithTracking(ctx, input)
		if err != nil {
			return nil, err
		}
	} else {
		response, err = a.runLocalWithTracking(ctx, input)
		if err != nil {
			return nil, err
		}
	}

	tracker.setExecutionTime(time.Since(startTime).Milliseconds())
	usage, execSummary, primaryModel := tracker.getResults()

	var execSum contracts.ExecutionSummary
	if execSummary != nil {
		execSum = *execSummary
	}

	// Log detailed execution information for all agent calls
	if detailed {
		executionDetails := map[string]interface{}{
			"agent_name":        a.name,
			"input_length":      len(input),
			"response_length":   len(response),
			"model_used":        primaryModel,
			"execution_time_ms": time.Since(startTime).Milliseconds(),
		}
		if execSummary != nil {
			executionDetails["llm_calls"] = execSummary.LLMCalls
			executionDetails["tool_calls"] = execSummary.ToolCalls
			executionDetails["sub_agent_calls"] = execSummary.SubAgentCalls
			executionDetails["used_tools"] = execSummary.UsedTools
			executionDetails["used_sub_agents"] = execSummary.UsedSubAgents
		}
		if usage != nil {
			executionDetails["input_tokens"] = usage.InputTokens
			executionDetails["output_tokens"] = usage.OutputTokens
			executionDetails["total_tokens"] = usage.TotalTokens
			executionDetails["reasoning_tokens"] = usage.ReasoningTokens
		}
		log.Printf("[Agent SDK] Agent execution completed: %+v", executionDetails)
	}

	return &contracts.AgentResponse{
		Content:          response,
		Usage:            usage,
		AgentName:        a.name,
		Model:            primaryModel,
		ExecutionSummary: execSum,
		Metadata: map[string]interface{}{
			"agent_name":            a.name,
			"execution_timestamp":   startTime.Unix(),
			"execution_duration_ms": time.Since(startTime).Milliseconds(),
		},
	}, nil
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
		if err := a.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleUser,
			Content: input,
		}); err != nil {
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
			if err := a.memory.AddMessage(ctx, contracts.Message{
				Role:    contracts.RoleAssistant,
				Content: response,
			}); err != nil {
				return "", fmt.Errorf("failed to add role response to memory: %w", err)
			}
		}

		return response, nil
	}

	// Use pre-initialized tools (manual + MCP tools already combined during agent creation).
	// initializeMCPTools already populated a.tools, so re-collecting here can append duplicates;
	// always run the merged slice through deduplicateTools to defend against that and against
	// MCP servers re-listing tools they already exposed at startup.
	allTools := a.tools

	if len(a.mcpServers) > 0 {
		mcpTools, err := a.collectMCPTools(ctx)
		if err != nil {
			// Log warning but continue - MCP tools are optional
			a.logger.Warn(context.Background(), fmt.Sprintf("Failed to collect MCP tools: %v", err), nil)
		} else if len(mcpTools) > 0 {
			allTools = deduplicateTools(append(allTools, mcpTools...))
		}
	}

	if len(a.lazyMCPConfigs) > 0 {
		lazyMCPTools := a.createLazyMCPTools()
		allTools = deduplicateTools(append(allTools, lazyMCPTools...))
	}

	if (len(allTools) > 0) && a.requirePlanApproval {
		a.planGenerator = plans.NewExecutionPlanGenerator(a.llm, allTools, a.systemPrompt, a.requirePlanApproval)
		return a.runWithExecutionPlan(ctx, input)
	}

	return a.runWithoutExecutionPlanWithToolsTracked(ctx, input, allTools)
}

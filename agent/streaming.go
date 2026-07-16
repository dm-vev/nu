package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dm-vev/nu/agent/execution"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
)

func sendEvent(ctx context.Context, eventChan chan<- contracts.AgentStreamEvent, event contracts.AgentStreamEvent) bool {
	select {
	case eventChan <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// RunStream executes the agent with a streaming response.
func (a *Agent) RunStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	if a.customRunStreamFunc != nil {
		return a.customRunStreamFunc(ctx, input, a)
	}
	if a.isRemote {
		return a.runRemoteStream(ctx, input)
	}
	return a.runLocalStream(ctx, input)
}

func (a *Agent) runLocalStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	streamingLLM, ok := a.llm.(contracts.StreamingLLM)
	if !ok {
		return nil, fmt.Errorf("LLM '%s' does not support streaming", a.llm.Name())
	}

	eventChan := make(chan contracts.AgentStreamEvent, 100)
	go a.runLocalStreamSession(ctx, input, streamingLLM, eventChan)
	return eventChan, nil
}

func (a *Agent) runLocalStreamSession(ctx context.Context, input string, streamingLLM contracts.StreamingLLM, eventChan chan<- contracts.AgentStreamEvent) {
	defer close(eventChan)

	startTime := time.Now()
	ctx = telemetry.WithAgentName(ctx, a.name)
	if a.orgID != "" {
		ctx = multitenancy.WithOrgID(ctx, a.orgID)
	}

	tracker := execution.NewTracker(true)
	ctx = execution.WithTracker(ctx, tracker)
	ctx, finishTrace := a.startStreamTrace(ctx, input, startTime, tracker)
	var responseLength int64
	defer func() { finishTrace(responseLength) }()

	processedInput, handled := a.prepareStreamInput(ctx, input, eventChan)
	if handled {
		return
	}
	allTools, handled := a.prepareStreamTools(ctx, processedInput, eventChan)
	if handled {
		return
	}

	length, err := a.runStreamingGeneration(ctx, processedInput, allTools, streamingLLM, eventChan)
	responseLength = length
	if err != nil {
		sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
			Type:      contracts.AgentEventError,
			Error:     err,
			Timestamp: time.Now(),
		})
	}
}

func (a *Agent) prepareStreamInput(ctx context.Context, input string, eventChan chan<- contracts.AgentStreamEvent) (string, bool) {
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{Role: contracts.RoleUser, Content: input}); err != nil {
			sendStreamError(ctx, eventChan, fmt.Errorf("failed to add user message to memory: %w", err))
			return "", true
		}
	}

	processedInput := input
	if a.guardrails != nil {
		guardedInput, err := a.guardrails.ProcessInput(ctx, input)
		if err != nil {
			sendStreamError(ctx, eventChan, fmt.Errorf("guardrails error: %w", err))
			return "", true
		}
		processedInput = guardedInput
	}

	taskID, action, planInput := a.extractPlanAction(processedInput)
	if taskID != "" {
		result, err := a.handlePlanAction(ctx, taskID, action, planInput)
		if err != nil {
			sendStreamError(ctx, eventChan, err)
		} else {
			sendEvent(ctx, eventChan, contracts.AgentStreamEvent{Type: contracts.AgentEventContent, Content: result, Timestamp: time.Now()})
		}
		return "", true
	}

	if a.systemPrompt != "" && a.isAskingAboutRole(processedInput) {
		response := a.generateRoleResponse()
		if a.memory != nil {
			if err := a.memory.AddMessage(ctx, contracts.Message{Role: contracts.RoleAssistant, Content: response}); err != nil {
				sendStreamError(ctx, eventChan, fmt.Errorf("failed to add role response to memory: %w", err))
				return "", true
			}
		}
		sendEvent(ctx, eventChan, contracts.AgentStreamEvent{Type: contracts.AgentEventContent, Content: response, Timestamp: time.Now()})
		sendEvent(ctx, eventChan, contracts.AgentStreamEvent{Type: contracts.AgentEventComplete, Timestamp: time.Now()})
		return "", true
	}
	return processedInput, false
}

func (a *Agent) prepareStreamTools(ctx context.Context, input string, eventChan chan<- contracts.AgentStreamEvent) ([]contracts.Tool, bool) {
	allTools := a.tools
	if len(a.mcpServers) > 0 {
		mcpTools, err := a.collectMCPTools(ctx)
		if err != nil {
			a.logger.Warn(ctx, "Failed to collect MCP tools", map[string]interface{}{"error": err.Error()})
		} else if len(mcpTools) > 0 {
			allTools = deduplicateTools(append(allTools, mcpTools...))
		}
	}
	if len(allTools) == 0 || !a.requirePlanApproval {
		return allTools, false
	}

	result, err := a.runWithExecutionPlan(ctx, input)
	if err != nil {
		sendStreamError(ctx, eventChan, err)
	} else {
		sendEvent(ctx, eventChan, contracts.AgentStreamEvent{Type: contracts.AgentEventContent, Content: result, Timestamp: time.Now()})
		sendEvent(ctx, eventChan, contracts.AgentStreamEvent{Type: contracts.AgentEventComplete, Timestamp: time.Now()})
	}
	return nil, true
}

func sendStreamError(ctx context.Context, eventChan chan<- contracts.AgentStreamEvent, err error) {
	sendEvent(ctx, eventChan, contracts.AgentStreamEvent{Type: contracts.AgentEventError, Error: err, Timestamp: time.Now()})
}

func (a *Agent) startStreamTrace(ctx context.Context, input string, startTime time.Time, tracker *execution.Tracker) (context.Context, func(int64)) {
	if a.tracer == nil {
		return ctx, func(int64) {}
	}

	ctx, span := a.tracer.StartSpan(ctx, "agent.RunStream")
	return ctx, func(responseLength int64) {
		if span == nil {
			return
		}
		tracker.SetExecutionTime(time.Since(startTime).Milliseconds())
		usage, summary, model := tracker.Results()
		spanData := map[string]interface{}{
			"agent_name":        a.name,
			"execution_time_ms": time.Since(startTime).Milliseconds(),
			"input_length":      len(input),
			"response_length":   responseLength,
		}
		if orgID, err := multitenancy.GetOrgID(ctx); err == nil && orgID != "" {
			spanData["org_id"] = orgID
		}
		if conversationID, ok := conversation.GetConversationID(ctx); ok && conversationID != "" {
			spanData["conversation_id"] = conversationID
		}
		if usage != nil {
			spanData["input_tokens"] = usage.InputTokens
			spanData["output_tokens"] = usage.OutputTokens
			spanData["total_tokens"] = usage.TotalTokens
			spanData["reasoning_tokens"] = usage.ReasoningTokens
		}
		if summary != nil {
			spanData["llm_calls"] = summary.LLMCalls
			spanData["tool_calls"] = summary.ToolCalls
			spanData["sub_agent_calls"] = summary.SubAgentCalls
			spanData["used_tools"] = summary.UsedTools
			spanData["used_sub_agents"] = summary.UsedSubAgents
		}
		if model != "" {
			spanData["model_used"] = model
		} else if a.llm != nil {
			spanData["model_used"] = a.llm.Name()
		}
		log.Printf("[Agent] RunStream execution completed: %+v", spanData)
		span.End()
	}
}

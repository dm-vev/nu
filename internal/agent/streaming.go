package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"nu/internal/contracts"
	"nu/internal/memory"
	"nu/internal/multitenancy"
	"nu/internal/telemetry"
)

// sendEvent pushes an AgentStreamEvent onto eventChan while respecting
// caller cancellation. Every blocking send on eventChan in this file goes
// through this helper so that abandoning the returned channel (timeout,
// client disconnect, etc.) doesn't leak the producing goroutine waiting
// on an unread channel (#291). Returns true on success, false if ctx was
// cancelled before the event could be delivered.
func sendEvent(ctx context.Context, eventChan chan<- contracts.AgentStreamEvent, event contracts.AgentStreamEvent) bool {
	select {
	case eventChan <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// RunStream executes the agent with streaming response
func (a *Agent) RunStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	// If custom stream function is set, use it instead
	if a.customRunStreamFunc != nil {
		return a.customRunStreamFunc(ctx, input, a)
	}

	// If this is a remote agent, delegate to remote execution
	if a.isRemote {
		return a.runRemoteStream(ctx, input)
	}

	// Local agent execution
	return a.runLocalStream(ctx, input)
}

// runLocalStream executes a local agent with streaming
func (a *Agent) runLocalStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	// Check if LLM supports streaming
	streamingLLM, ok := a.llm.(contracts.StreamingLLM)
	if !ok {
		return nil, fmt.Errorf("LLM '%s' does not support streaming", a.llm.Name())
	}

	// Get buffer size from default config
	bufferSize := 100

	// Create agent event channel
	eventChan := make(chan contracts.AgentStreamEvent, bufferSize)

	// Start streaming in a goroutine
	go func() {
		defer close(eventChan)

		// Track execution start time
		startTime := time.Now()

		// Inject agent name into context for tracing span naming
		ctx = telemetry.WithAgentName(ctx, a.name)

		// If orgID is set on the agent, add it to the context
		if a.orgID != "" {
			ctx = multitenancy.WithOrgID(ctx, a.orgID)
		}

		// Create usage tracker for detailed metrics collection
		tracker := newUsageTracker(true)
		ctx = withUsageTracker(ctx, tracker)

		// Track response length for span logging
		var responseLength int64

		// Start tracing if available
		var span contracts.Span
		if a.tracer != nil {
			ctx, span = a.tracer.StartSpan(ctx, "agent.RunStream")
			defer func() {
				// Add detailed execution information to span before ending
				if span != nil {
					executionTimeMs := time.Since(startTime).Milliseconds()
					tracker.setExecutionTime(executionTimeMs)

					usage, execSummary, model := tracker.getResults()

					// Add comprehensive span attributes
					spanData := map[string]interface{}{
						"agent_name":        a.name,
						"execution_time_ms": executionTimeMs,
						"input_length":      len(input),
						"response_length":   responseLength,
					}

					// Add organization and conversation context if available
					if orgID, err := multitenancy.GetOrgID(ctx); err == nil && orgID != "" {
						spanData["org_id"] = orgID
					}
					if conversationID, ok := memory.GetConversationID(ctx); ok && conversationID != "" {
						spanData["conversation_id"] = conversationID
					}

					// Add token usage if available
					if usage != nil {
						spanData["input_tokens"] = usage.InputTokens
						spanData["output_tokens"] = usage.OutputTokens
						spanData["total_tokens"] = usage.TotalTokens
						spanData["reasoning_tokens"] = usage.ReasoningTokens
					}

					// Add execution summary if available
					if execSummary != nil {
						spanData["llm_calls"] = execSummary.LLMCalls
						spanData["tool_calls"] = execSummary.ToolCalls
						spanData["sub_agent_calls"] = execSummary.SubAgentCalls
						spanData["used_tools"] = execSummary.UsedTools
						spanData["used_sub_agents"] = execSummary.UsedSubAgents
					}

					// Add model information
					if model != "" {
						spanData["model_used"] = model
					} else if a.llm != nil {
						spanData["model_used"] = a.llm.Name()
					}

					// Log detailed execution information
					log.Printf("[Agent] RunStream execution completed: %+v", spanData)
				}
				span.End()
			}()
		}

		// Add user message to memory
		if a.memory != nil {
			if err := a.memory.AddMessage(ctx, contracts.Message{
				Role:    "user",
				Content: input,
			}); err != nil {
				sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("failed to add user message to memory: %w", err),
					Timestamp: time.Now(),
				})
				return
			}
		}

		// Apply guardrails to input if available
		processedInput := input
		if a.guardrails != nil {
			guardedInput, err := a.guardrails.ProcessInput(ctx, input)
			if err != nil {
				sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("guardrails error: %w", err),
					Timestamp: time.Now(),
				})
				return
			}
			processedInput = guardedInput
		}

		// Check if the input is related to an existing plan
		taskID, action, planInput := a.extractPlanAction(processedInput)
		if taskID != "" {
			// For now, plan actions are not streamed - fall back to regular handling
			result, err := a.handlePlanAction(ctx, taskID, action, planInput)
			if err != nil {
				sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     err,
					Timestamp: time.Now(),
				})
			} else {
				sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
					Type:      contracts.AgentEventContent,
					Content:   result,
					Timestamp: time.Now(),
				})
			}
			return
		}

		// Check if the user is asking about the agent's role or identity
		if a.systemPrompt != "" && a.isAskingAboutRole(processedInput) {
			response := a.generateRoleResponse()

			// Add the role response to memory if available
			if a.memory != nil {
				if err := a.memory.AddMessage(ctx, contracts.Message{
					Role:    "assistant",
					Content: response,
				}); err != nil {
					sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
						Type:      contracts.AgentEventError,
						Error:     fmt.Errorf("failed to add role response to memory: %w", err),
						Timestamp: time.Now(),
					})
					return
				}
			}

			sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
				Type:      contracts.AgentEventContent,
				Content:   response,
				Timestamp: time.Now(),
			})
			sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
				Type:      contracts.AgentEventComplete,
				Timestamp: time.Now(),
			})
			return
		}

		// Collect all tools. initializeMCPTools already populated a.tools, so the
		// runtime re-collect below can re-add the same tools; deduplicate after the
		// append to keep tool names unique (LLM providers like Anthropic reject
		// requests with duplicate tool names — see issue #308).
		allTools := a.tools

		// Add MCP tools if available
		if len(a.mcpServers) > 0 {
			mcpTools, err := a.collectMCPTools(ctx)
			if err != nil {
				// Log the error but continue with the agent tools
				// Warning: Failed to collect MCP tools
				a.logger.Warn(ctx, "Failed to collect MCP tools", map[string]interface{}{"error": err.Error()})
			} else if len(mcpTools) > 0 {
				allTools = deduplicateTools(append(allTools, mcpTools...))
			}
		}

		// If tools are available and plan approval is required, we can't stream execution plans yet
		if (len(allTools) > 0) && a.requirePlanApproval {
			// For now, fall back to non-streaming execution plan generation
			result, err := a.runWithExecutionPlan(ctx, processedInput)
			if err != nil {
				sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     err,
					Timestamp: time.Now(),
				})
			} else {
				sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
					Type:      contracts.AgentEventContent,
					Content:   result,
					Timestamp: time.Now(),
				})
				sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
					Type:      contracts.AgentEventComplete,
					Timestamp: time.Now(),
				})
			}
			return
		}

		// Run with streaming
		length, err := a.runStreamingGeneration(ctx, processedInput, allTools, streamingLLM, eventChan)
		responseLength = length
		if err != nil {
			sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
				Type:      contracts.AgentEventError,
				Error:     err,
				Timestamp: time.Now(),
			})
		}
	}()

	return eventChan, nil
}

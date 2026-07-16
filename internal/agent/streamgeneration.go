package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nu/internal/contracts"
)

// runStreamingGeneration handles the core streaming generation logic
func (a *Agent) runStreamingGeneration(
	ctx context.Context,
	input string,
	allTools []contracts.Tool,
	streamingLLM contracts.StreamingLLM,
	eventChan chan<- contracts.AgentStreamEvent,
) (int64, error) {
	// Prepare generation options
	options := []contracts.GenerateOption{}

	// Add system prompt if available
	if a.systemPrompt != "" {
		options = append(options, func(opts *contracts.GenerateOptions) {
			opts.SystemMessage = a.systemPrompt
		})
	}

	// Add LLM config if available
	if a.llmConfig != nil {
		options = append(options, func(opts *contracts.GenerateOptions) {
			opts.LLMConfig = a.llmConfig
		})
	}

	// Add response format if available
	if a.responseFormat != nil {
		options = append(options, func(opts *contracts.GenerateOptions) {
			opts.ResponseFormat = a.responseFormat
		})
	}

	// Add max iterations if available
	if a.maxIterations > 0 {
		options = append(options, contracts.WithMaxIterations(a.maxIterations))
	}

	// Add memory if available
	if a.memory != nil {
		options = append(options, contracts.WithMemory(a.memory))
	}

	// Add stream config if available
	if a.streamConfig != nil {
		options = append(options, contracts.WithStreamConfig(*a.streamConfig))
	}

	// Add cache config if available
	if a.cacheConfig != nil {
		options = append(options, func(opts *contracts.GenerateOptions) {
			opts.CacheConfig = a.cacheConfig
		})
	}

	// Inject stream forwarder into context so sub-agents can forward their events
	// This allows nested sub-agent streaming to work properly
	streamForwarder := func(event contracts.AgentStreamEvent) {
		// Forward sub-agent events to the parent agent's event channel
		select {
		case eventChan <- event:
		case <-ctx.Done():
			// Context cancelled, stop forwarding
		}
	}

	// Add the stream forwarder to context
	// This is used by the tools package's AgentTool to forward sub-agent events
	ctxWithForwarder := context.WithValue(ctx, contracts.StreamForwarderKey, contracts.StreamForwarder(streamForwarder))

	// Start LLM streaming
	var llmEventChan <-chan contracts.StreamEvent
	var err error

	if len(allTools) > 0 {
		// Record tool invocations as the LLM actually calls them, not the
		// full set of available tools (#305).
		toolsForLLM := wrapToolsWithTracker(allTools, getUsageTracker(ctx))
		llmEventChan, err = streamingLLM.GenerateWithToolsStream(ctxWithForwarder, input, toolsForLLM, options...)
	} else {
		llmEventChan, err = streamingLLM.GenerateStream(ctxWithForwarder, input, options...)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to start LLM streaming: %w", err)
	}

	// Track accumulated content and tool calls for memory
	var accumulatedContent strings.Builder
	var toolCalls []contracts.ToolCall
	var toolResults map[string]string // map[toolCallID]result
	var finalError error

	toolResults = make(map[string]string)

	// Forward LLM events as agent events
	for llmEvent := range llmEventChan {
		agentEvent := a.convertLLMEventToAgentEvent(llmEvent, allTools)

		// Accumulate content for memory (not thinking)
		if llmEvent.Type == contracts.StreamEventContentDelta {
			accumulatedContent.WriteString(llmEvent.Content)
		}

		// Track tool calls for memory
		if llmEvent.Type == contracts.StreamEventToolUse && llmEvent.ToolCall != nil {
			toolCalls = append(toolCalls, *llmEvent.ToolCall)
		}

		// Track tool results for memory
		if llmEvent.Type == contracts.StreamEventToolResult && llmEvent.ToolCall != nil {
			toolResults[llmEvent.ToolCall.ID] = llmEvent.Content
		}

		// Track errors
		if llmEvent.Error != nil {
			finalError = llmEvent.Error
		}

		// Send agent event
		if !sendEvent(ctx, eventChan, agentEvent) {
			return int64(accumulatedContent.Len()), finalError
		}
	}

	// Add messages to memory if available (save even on error to preserve conversation history)
	if a.memory != nil {
		// If we have tool calls, save them in the correct order
		if len(toolCalls) > 0 {
			// Add assistant message with tool calls
			err := a.memory.AddMessage(ctx, contracts.Message{
				Role:      "assistant",
				Content:   accumulatedContent.String(), // May be empty or contain text before tools
				ToolCalls: toolCalls,
			})
			if err != nil {
				a.logger.Warn(ctx, "Failed to add assistant tool calls to memory", map[string]interface{}{"error": err.Error()})
			}

			// Add tool result messages
			for _, toolCall := range toolCalls {
				if result, ok := toolResults[toolCall.ID]; ok {
					err := a.memory.AddMessage(ctx, contracts.Message{
						Role:       "tool",
						Content:    result,
						ToolCallID: toolCall.ID,
						Metadata: map[string]interface{}{
							"tool_name": toolCall.Name,
						},
					})
					if err != nil {
						a.logger.Warn(ctx, "Failed to add tool result to memory", map[string]interface{}{"error": err.Error()})
					}
				}
			}
		} else if accumulatedContent.Len() > 0 {
			// No tool calls, just content - add assistant message
			err := a.memory.AddMessage(ctx, contracts.Message{
				Role:    "assistant",
				Content: accumulatedContent.String(),
			})
			if err != nil {
				a.logger.Warn(ctx, "Failed to add assistant response to memory", map[string]interface{}{"error": err.Error()})
			}
		}
	}

	// Send completion event
	sendEvent(ctx, eventChan, contracts.AgentStreamEvent{
		Type:      contracts.AgentEventComplete,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"total_content_length": accumulatedContent.Len(),
			"had_error":            finalError != nil,
		},
	})

	return int64(accumulatedContent.Len()), finalError
}

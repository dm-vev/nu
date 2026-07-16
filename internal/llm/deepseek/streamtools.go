package deepseek

import (
	"context"
	"time"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// GenerateWithToolsStream implements contracts.StreamingLLM.GenerateWithToolsStream with iterative tool calling
func (c *Client) GenerateWithToolsStream(
	ctx context.Context,
	prompt string,
	tools []contracts.Tool,
	options ...contracts.GenerateOption,
) (<-chan contracts.StreamEvent, error) {
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
		maxIterations = 2
	}

	// Check for organization ID in context
	defaultOrgID := "default"
	if id, err := multitenancy.GetOrgID(ctx); err == nil {
		ctx = multitenancy.WithOrgID(ctx, id)
	} else {
		ctx = multitenancy.WithOrgID(ctx, defaultOrgID)
	}

	// Get buffer size from stream config
	bufferSize := 100
	if params.StreamConfig != nil {
		bufferSize = params.StreamConfig.BufferSize
	}

	// Create event channel
	eventChan := make(chan contracts.StreamEvent, bufferSize)

	// Start streaming with iterative tool calling
	go c.runToolsStream(ctx, prompt, tools, params, maxIterations, eventChan)

	return eventChan, nil
}

func (c *Client) runToolsStream(
	ctx context.Context,
	prompt string,
	tools []contracts.Tool,
	params *contracts.GenerateOptions,
	maxIterations int,
	eventChan chan contracts.StreamEvent,
) {
	defer close(eventChan)

	// Convert tools to DeepSeek format
	deepseekTools := c.convertToolsToDeepSeekFormat(tools)

	// Build messages
	messages := []Message{}

	// Add system message if provided
	if params.SystemMessage != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: params.SystemMessage,
		})
		c.logger.Debug(ctx, "Using system message for tools", map[string]interface{}{"system_message": params.SystemMessage})
	}

	// Build messages using unified builder
	builder := deepSeekNewMessageHistoryBuilder(c.logger)
	messages = append(messages, builder.buildMessages(ctx, prompt, params.Memory)...)

	// Send initial message start event
	eventChan <- contracts.StreamEvent{
		Type:      contracts.StreamEventMessageStart,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"model": c.Model,
			"tools": len(deepseekTools),
		},
	}

	// Determine if we should filter intermediate content
	filterIntermediateContent := params.StreamConfig == nil || !params.StreamConfig.IncludeIntermediateMessages

	// Track captured content for final iteration replay if filtering is enabled
	var capturedContentEvents []contracts.StreamEvent

	// Track if we got a complete response (no tool calls)
	gotCompleteResponse := false

	// Iterative tool calling loop
	for iteration := 0; iteration < maxIterations; iteration++ {
		result, ok := c.streamToolsIteration(
			ctx, params, deepseekTools, messages, iteration, maxIterations,
			filterIntermediateContent, eventChan,
		)
		if !ok {
			return
		}

		// Check if the model wants to use tools
		if len(result.assistantResponse.ToolCalls) == 0 {
			// No tool calls, we're done
			if result.hasContent {
				eventChan <- contracts.StreamEvent{
					Type:      contracts.StreamEventContentComplete,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"iteration": iteration + 1,
					},
				}
			}
			gotCompleteResponse = true
			break
		}

		// The model wants to use tools
		c.logger.Info(ctx, "Processing tool calls", map[string]interface{}{
			"count":     len(result.assistantResponse.ToolCalls),
			"iteration": iteration + 1,
		})

		// Add the assistant's message with tool calls to the conversation
		messages = append(messages, result.assistantResponse)
		messages = append(messages, c.executeStreamTools(ctx, tools, result.assistantResponse.ToolCalls, iteration, eventChan)...)

		// If we had content during this iteration and tools were called, capture it for final replay
		if filterIntermediateContent && result.hasContent {
			capturedContentEvents = append(capturedContentEvents, result.contentEvents...)
		}
	}

	// Replay captured content events if we filtered them during iterations
	if filterIntermediateContent && len(capturedContentEvents) > 0 {
		c.logger.Debug(ctx, "Replaying captured content events from tool iterations", map[string]interface{}{
			"eventsCount": len(capturedContentEvents),
		})
		for _, contentEvent := range capturedContentEvents {
			eventChan <- contentEvent
		}
	}

	// If we got a complete response (no tool calls), skip the final synthesis call
	if gotCompleteResponse {
		c.logger.Debug(ctx, "Skipping final synthesis call - already got complete response", map[string]interface{}{
			"maxIterations": maxIterations,
		})
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStop,
			Timestamp: time.Now(),
		}
		return
	}

	// If DisableFinalSummary is enabled, skip the final synthesis call
	if params.DisableFinalSummary {
		c.logger.Info(ctx, "DisableFinalSummary enabled, skipping final synthesis call", map[string]interface{}{
			"maxIterations": maxIterations,
		})
		eventChan <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStop,
			Timestamp: time.Now(),
		}
		return
	}

	c.streamFinalSynthesis(ctx, params, messages, maxIterations, eventChan)
}

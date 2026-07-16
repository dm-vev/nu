package gemini

import (
	"context"
	"time"

	"nu/internal/contracts"
)

// GenerateWithToolsStream generates text with tools and streaming response with real-time tool events
func (c *Client) GenerateWithToolsStream(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	// Convert options to params
	params := &contracts.GenerateOptions{}
	for _, opt := range options {
		if opt != nil {
			opt(params)
		}
	}

	// Set default values only if they're not provided
	if params.LLMConfig == nil {
		params.LLMConfig = &contracts.LLMConfig{
			Temperature:      0.7,
			TopP:             1.0,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
		}
	}

	// Set default max iterations if not provided
	maxIterations := params.MaxIterations
	if maxIterations == 0 {
		maxIterations = 2 // Default to current behavior
	}

	// Get streaming config or use default
	streamConfig := contracts.DefaultStreamConfig()
	if params.StreamConfig != nil {
		streamConfig = *params.StreamConfig
	}

	// Create event channel
	eventCh := make(chan contracts.StreamEvent, streamConfig.BufferSize)

	go func() {
		defer close(eventCh)

		// Send message start event
		select {
		case eventCh <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStart,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}

		c.logger.Debug(ctx, "Starting streaming with tools with real-time events", map[string]interface{}{
			"model":         c.model,
			"tools":         len(tools),
			"maxIterations": maxIterations,
		})

		// Execute the tool calling process with streaming events
		response, err := c.generateWithToolsAndStream(ctx, prompt, tools, params, maxIterations, eventCh)
		if err != nil {
			// Send error event
			select {
			case eventCh <- contracts.StreamEvent{
				Type:      contracts.StreamEventError,
				Error:     err,
				Timestamp: time.Now(),
			}:
			case <-ctx.Done():
			}
			return
		}

		// Stream the final response in chunks
		c.streamResponse(ctx, response, eventCh)

		// Send content complete event
		select {
		case eventCh <- contracts.StreamEvent{
			Type:      contracts.StreamEventContentComplete,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}

		// Send message stop event
		select {
		case eventCh <- contracts.StreamEvent{
			Type:      contracts.StreamEventMessageStop,
			Timestamp: time.Now(),
		}:
		case <-ctx.Done():
			return
		}

		c.logger.Info(ctx, "Successfully completed streaming response with tools", map[string]interface{}{
			"maxIterations": maxIterations,
		})
	}()

	return eventCh, nil
}

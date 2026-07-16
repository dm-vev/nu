package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/contracts"
)

// runWithStreaming runs the sub-agent with streaming and forwards events to the parent
func (at *AgentTool) runWithStreaming(ctx context.Context, input string, forwarder contracts.StreamForwarder, span contracts.Span, agentName string) (string, error) {
	// Start streaming from the sub-agent
	eventChan, err := at.agent.RunStream(ctx, input)
	if err != nil {
		at.logger.Error(ctx, "Failed to start sub-agent streaming", map[string]interface{}{
			"sub_agent": agentName,
			"error":     err.Error(),
		})
		return "", fmt.Errorf("failed to start sub-agent streaming: %w", err)
	}

	// Log that we're streaming
	at.logger.Debug(ctx, "Sub-agent streaming started", map[string]interface{}{
		"sub_agent": agentName,
		"tool_name": at.name,
	})

	// Collect content for final result
	var contentBuilder strings.Builder
	var finalError error

	// Forward all events and collect content
	for event := range eventChan {
		// Forward the event to the parent stream
		forwarder(event)

		// Collect content for the final result
		if event.Type == contracts.AgentEventContent {
			contentBuilder.WriteString(event.Content)
		}

		// Track errors
		if event.Error != nil {
			finalError = event.Error
			at.logger.Error(ctx, "Sub-agent streaming error", map[string]interface{}{
				"sub_agent": agentName,
				"error":     event.Error.Error(),
			})
		}

		// Add event to span if available
		if span != nil {
			span.AddEvent(fmt.Sprintf("sub_agent_%s", event.Type), map[string]interface{}{
				"type":      string(event.Type),
				"sub_agent": agentName,
				"has_error": event.Error != nil,
			})
		}
	}

	at.logger.Debug(ctx, "Sub-agent streaming completed", map[string]interface{}{
		"sub_agent":    agentName,
		"tool_name":    at.name,
		"response_len": contentBuilder.Len(),
	})

	// Return error if we encountered one
	if finalError != nil {
		return "", finalError
	}

	return contentBuilder.String(), nil
}

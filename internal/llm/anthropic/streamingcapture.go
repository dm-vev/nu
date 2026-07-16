package anthropic

import (
	"context"
	"time"

	"nu/internal/contracts"
)

func (c *Client) createFilteredEventForwarder(ctx context.Context, tempEventChan <-chan contracts.StreamEvent, eventChan chan<- contracts.StreamEvent, filterContentDeltas bool) ([]contracts.ToolCall, bool, []contracts.StreamEvent, error) {
	var toolCalls []contracts.ToolCall
	var hasContent bool
	var capturedContentEvents []contracts.StreamEvent
	for event := range tempEventChan {
		if event.Type == contracts.StreamEventContentDelta && event.Content != "" {
			hasContent = true
			capturedContentEvents = append(capturedContentEvents, event)
			if filterContentDeltas {
				continue
			}
		}
		select {
		case eventChan <- event:
		case <-ctx.Done():
			return nil, false, nil, ctx.Err()
		}
		if event.Type == contracts.StreamEventToolUse && event.ToolCall != nil {
			toolCalls = append(toolCalls, *event.ToolCall)
		}
		if event.Error != nil {
			return nil, false, nil, event.Error
		}
	}
	return toolCalls, hasContent, capturedContentEvents, nil
}

func (c *Client) executeStreamingRequestWithToolCapture(ctx context.Context, req CompletionRequest, eventChan chan<- contracts.StreamEvent, filterContentDeltas bool, params *contracts.GenerateOptions) ([]contracts.ToolCall, bool, []contracts.StreamEvent, error) {
	tempEventChan := make(chan contracts.StreamEvent, 100)
	go func() {
		defer func() {
			defer func() { _ = recover() }()
			close(tempEventChan)
		}()
		if err := c.executeStreamingRequestWithMemory(ctx, req, tempEventChan, "", params); err != nil {
			select {
			case tempEventChan <- contracts.StreamEvent{Type: contracts.StreamEventError, Error: err, Timestamp: time.Now()}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return c.createFilteredEventForwarder(ctx, tempEventChan, eventChan, filterContentDeltas)
}

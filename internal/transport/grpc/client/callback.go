package client

import (
	"context"

	"github.com/dm-vev/nu/contracts"
)

// OnThinking registers a handler for thinking events
func (r *Client) OnThinking(handler func(string)) *Client {
	r.handlersMu.Lock()
	defer r.handlersMu.Unlock()
	r.thinkingHandlers = append(r.thinkingHandlers, handler)
	return r
}

// OnContent registers a handler for content events
func (r *Client) OnContent(handler func(string)) *Client {
	r.handlersMu.Lock()
	defer r.handlersMu.Unlock()
	r.contentHandlers = append(r.contentHandlers, handler)
	return r
}

// OnToolCall registers a handler for tool call events
func (r *Client) OnToolCall(handler func(*contracts.ToolCallEvent)) *Client {
	r.handlersMu.Lock()
	defer r.handlersMu.Unlock()
	r.toolCallHandlers = append(r.toolCallHandlers, handler)
	return r
}

// OnToolResult registers a handler for tool result events
func (r *Client) OnToolResult(handler func(*contracts.ToolCallEvent)) *Client {
	r.handlersMu.Lock()
	defer r.handlersMu.Unlock()
	r.toolResultHandlers = append(r.toolResultHandlers, handler)
	return r
}

// OnError registers a handler for error events
func (r *Client) OnError(handler func(error)) *Client {
	r.handlersMu.Lock()
	defer r.handlersMu.Unlock()
	r.errorHandlers = append(r.errorHandlers, handler)
	return r
}

// OnComplete registers a handler for completion events
func (r *Client) OnComplete(handler func()) *Client {
	r.handlersMu.Lock()
	defer r.handlersMu.Unlock()
	r.completeHandlers = append(r.completeHandlers, handler)
	return r
}

// Stream executes the remote agent with registered event handlers
func (r *Client) Stream(ctx context.Context, input string) error {
	events, err := r.RunStream(ctx, input)
	if err != nil {
		return err
	}

	for event := range events {
		// Execute handlers synchronously to preserve event ordering
		r.executeHandlers(event)
	}

	return nil
}

// StreamWithAuth executes the remote agent with registered event handlers and explicit auth token
func (r *Client) StreamWithAuth(ctx context.Context, input string, authToken string) error {
	events, err := r.RunStreamWithAuth(ctx, input, authToken)
	if err != nil {
		return err
	}

	for event := range events {
		// Execute handlers synchronously to preserve event ordering
		r.executeHandlers(event)
	}

	return nil
}

// executeHandlers executes the appropriate handlers for an event
func (r *Client) executeHandlers(event contracts.AgentStreamEvent) {
	r.handlersMu.RLock()
	defer r.handlersMu.RUnlock()

	switch event.Type {
	case contracts.AgentEventThinking:
		for _, handler := range r.thinkingHandlers {
			handler(event.ThinkingStep)
		}

	case contracts.AgentEventContent:
		for _, handler := range r.contentHandlers {
			handler(event.Content)
		}

	case contracts.AgentEventToolCall:
		for _, handler := range r.toolCallHandlers {
			handler(event.ToolCall)
		}

	case contracts.AgentEventToolResult:
		for _, handler := range r.toolResultHandlers {
			handler(event.ToolCall)
		}

	case contracts.AgentEventError:
		for _, handler := range r.errorHandlers {
			handler(event.Error)
		}

	case contracts.AgentEventComplete:
		for _, handler := range r.completeHandlers {
			handler()
		}
	}
}

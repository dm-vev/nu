package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"nu/internal/contracts"
	pb "nu/internal/transport/grpc/pb"
)

// RunStream executes the agent with streaming response
func (m *Microservice) RunStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	if !m.IsRunning() {
		return nil, fmt.Errorf("microservice is not running")
	}

	// Create gRPC client
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", m.port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to microservice: %w", err)
	}

	client := pb.NewAgentServiceClient(conn)
	stream, err := client.RunStream(ctx, &pb.RunRequest{
		Input: input,
	})
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	// Create output channel
	outputCh := make(chan contracts.AgentStreamEvent, 100)

	// Start goroutine to process stream
	go func() {
		defer func() {
			_ = conn.Close()
			close(outputCh)
		}()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					return
				}
				// Send error event
				select {
				case outputCh <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     err,
					Timestamp: time.Now(),
				}:
				case <-ctx.Done():
				}
				return
			}

			// Convert gRPC response to AgentStreamEvent
			event := convertGRPCResponseToAgentEvent(response)

			// Send event to channel
			select {
			case outputCh <- event:
			case <-ctx.Done():
				return
			}

			// Check if final
			if response.IsFinal || response.EventType == pb.EventType_EVENT_TYPE_COMPLETE {
				return
			}
		}
	}()

	return outputCh, nil
}

// OnThinking registers a handler for thinking events
func (m *Microservice) OnThinking(handler func(string)) *Microservice {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.thinkingHandlers = append(m.thinkingHandlers, handler)
	return m
}

// OnContent registers a handler for content events
func (m *Microservice) OnContent(handler func(string)) *Microservice {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.contentHandlers = append(m.contentHandlers, handler)
	return m
}

// OnToolCall registers a handler for tool call events
func (m *Microservice) OnToolCall(handler func(*contracts.ToolCallEvent)) *Microservice {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.toolCallHandlers = append(m.toolCallHandlers, handler)
	return m
}

// OnToolResult registers a handler for tool result events
func (m *Microservice) OnToolResult(handler func(*contracts.ToolCallEvent)) *Microservice {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.toolResultHandlers = append(m.toolResultHandlers, handler)
	return m
}

// OnError registers a handler for error events
func (m *Microservice) OnError(handler func(error)) *Microservice {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.errorHandlers = append(m.errorHandlers, handler)
	return m
}

// OnComplete registers a handler for completion events
func (m *Microservice) OnComplete(handler func()) *Microservice {
	m.handlersMu.Lock()
	defer m.handlersMu.Unlock()
	m.completeHandlers = append(m.completeHandlers, handler)
	return m
}

// Stream executes the agent with registered event handlers
func (m *Microservice) Stream(ctx context.Context, input string) error {
	events, err := m.RunStream(ctx, input)
	if err != nil {
		return err
	}

	for event := range events {
		m.executeHandlers(event)
	}

	return nil
}

// executeHandlers executes the appropriate handlers for an event
func (m *Microservice) executeHandlers(event contracts.AgentStreamEvent) {
	m.handlersMu.RLock()
	defer m.handlersMu.RUnlock()

	switch event.Type {
	case contracts.AgentEventThinking:
		for _, handler := range m.thinkingHandlers {
			handler(event.ThinkingStep)
		}

	case contracts.AgentEventContent:
		for _, handler := range m.contentHandlers {
			handler(event.Content)
		}

	case contracts.AgentEventToolCall:
		for _, handler := range m.toolCallHandlers {
			handler(event.ToolCall)
		}

	case contracts.AgentEventToolResult:
		for _, handler := range m.toolResultHandlers {
			handler(event.ToolCall)
		}

	case contracts.AgentEventError:
		for _, handler := range m.errorHandlers {
			handler(event.Error)
		}

	case contracts.AgentEventComplete:
		for _, handler := range m.completeHandlers {
			handler()
		}
	}
}

// convertGRPCResponseToAgentEvent converts gRPC stream response to AgentStreamEvent
func convertGRPCResponseToAgentEvent(response *pb.RunStreamResponse) contracts.AgentStreamEvent {
	event := contracts.AgentStreamEvent{
		Timestamp: time.Now(),
	}

	// Set timestamp from response if available
	if response.Timestamp > 0 {
		event.Timestamp = time.UnixMilli(response.Timestamp)
	}

	// Convert metadata
	if response.Metadata != nil {
		event.Metadata = make(map[string]interface{})
		for k, v := range response.Metadata {
			event.Metadata[k] = v
		}
	}

	// Convert based on event type
	switch response.EventType {
	case pb.EventType_EVENT_TYPE_THINKING:
		event.Type = contracts.AgentEventThinking
		event.ThinkingStep = response.Thinking

	case pb.EventType_EVENT_TYPE_CONTENT:
		event.Type = contracts.AgentEventContent
		event.Content = response.Chunk

	case pb.EventType_EVENT_TYPE_TOOL_CALL, pb.EventType_EVENT_TYPE_TOOL_RESULT:
		if response.EventType == pb.EventType_EVENT_TYPE_TOOL_CALL {
			event.Type = contracts.AgentEventToolCall
		} else {
			event.Type = contracts.AgentEventToolResult
		}

		if response.ToolCall != nil {
			event.ToolCall = &contracts.ToolCallEvent{
				ID:          response.ToolCall.Id,
				Name:        response.ToolCall.Name,
				DisplayName: response.ToolCall.DisplayName,
				Internal:    response.ToolCall.Internal,
				Arguments:   response.ToolCall.Arguments,
				Result:      response.ToolCall.Result,
				Status:      response.ToolCall.Status,
			}
		}

	case pb.EventType_EVENT_TYPE_ERROR:
		event.Type = contracts.AgentEventError
		if response.Error != "" {
			event.Error = fmt.Errorf("%s", response.Error)
		}

	case pb.EventType_EVENT_TYPE_COMPLETE:
		event.Type = contracts.AgentEventComplete

	default:
		// Default to content for unknown event types
		event.Type = contracts.AgentEventContent
		event.Content = response.Chunk
	}

	return event
}

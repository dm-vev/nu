package grpc

import (
	"time"

	"nu/internal/contracts"
	pb "nu/internal/transport/grpc/pb"
)

// convertPbToStreamEvent converts a protobuf RunStreamResponse to AgentStreamEvent
func convertPbToStreamEvent(resp *pb.RunStreamResponse) contracts.AgentStreamEvent {
	event := contracts.AgentStreamEvent{
		Content:   resp.Chunk,
		Timestamp: time.Unix(resp.Timestamp, 0),
		Metadata:  make(map[string]interface{}),
	}

	// Copy metadata
	for k, v := range resp.Metadata {
		event.Metadata[k] = v
	}

	// Convert event type
	switch resp.EventType {
	case pb.EventType_EVENT_TYPE_CONTENT:
		event.Type = contracts.AgentEventContent
	case pb.EventType_EVENT_TYPE_THINKING:
		event.Type = contracts.AgentEventThinking
		event.ThinkingStep = resp.Thinking
	case pb.EventType_EVENT_TYPE_TOOL_CALL:
		event.Type = contracts.AgentEventToolCall
		if resp.ToolCall != nil {
			event.ToolCall = &contracts.ToolCallEvent{
				ID:          resp.ToolCall.Id,
				Name:        resp.ToolCall.Name,
				DisplayName: resp.ToolCall.DisplayName,
				Internal:    resp.ToolCall.Internal,
				Arguments:   resp.ToolCall.Arguments,
				Result:      resp.ToolCall.Result,
				Status:      resp.ToolCall.Status,
			}
		}
	case pb.EventType_EVENT_TYPE_TOOL_RESULT:
		event.Type = contracts.AgentEventToolResult
		if resp.ToolCall != nil {
			event.ToolCall = &contracts.ToolCallEvent{
				ID:          resp.ToolCall.Id,
				Name:        resp.ToolCall.Name,
				DisplayName: resp.ToolCall.DisplayName,
				Internal:    resp.ToolCall.Internal,
				Arguments:   resp.ToolCall.Arguments,
				Result:      resp.ToolCall.Result,
				Status:      resp.ToolCall.Status,
			}
		}
	case pb.EventType_EVENT_TYPE_ERROR:
		event.Type = contracts.AgentEventError
	case pb.EventType_EVENT_TYPE_COMPLETE:
		event.Type = contracts.AgentEventComplete
	default:
		event.Type = contracts.AgentEventContent
	}

	// Handle timestamp (use current time if not provided)
	if resp.Timestamp == 0 {
		event.Timestamp = time.Now()
	}

	return event
}

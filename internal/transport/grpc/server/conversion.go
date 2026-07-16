package server

import (
	"fmt"

	"nu/internal/contracts"
	pb "nu/internal/transport/grpc/pb"
)

// convertEventType converts agent event types to protobuf event types
func (s *Server) convertEventType(eventType contracts.AgentEventType) pb.EventType {
	switch eventType {
	case contracts.AgentEventContent:
		return pb.EventType_EVENT_TYPE_CONTENT
	case contracts.AgentEventThinking:
		return pb.EventType_EVENT_TYPE_THINKING
	case contracts.AgentEventToolCall:
		return pb.EventType_EVENT_TYPE_TOOL_CALL
	case contracts.AgentEventToolResult:
		return pb.EventType_EVENT_TYPE_TOOL_RESULT
	case contracts.AgentEventError:
		return pb.EventType_EVENT_TYPE_ERROR
	case contracts.AgentEventComplete:
		return pb.EventType_EVENT_TYPE_COMPLETE
	default:
		return pb.EventType_EVENT_TYPE_CONTENT
	}
}

// formatExecutionPlan formats an execution plan for display
func formatExecutionPlan(plan interface{}) string {
	// This is a placeholder implementation
	// In reality, you would use the actual execution plan formatting logic
	return fmt.Sprintf("Execution Plan: %+v", plan)
}

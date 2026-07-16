package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"nu/internal/contracts"
	"nu/internal/memory"
	"nu/internal/multitenancy"
	pb "nu/internal/transport/grpc/pb"
)

// Run executes the agent with the given input
func (s *AgentServer) Run(ctx context.Context, req *pb.RunRequest) (*pb.RunResponse, error) {
	if req.Input == "" {
		return nil, status.Error(codes.InvalidArgument, "input cannot be empty")
	}

	// Add org_id to context if provided
	if req.OrgId != "" {
		ctx = multitenancy.WithOrgID(ctx, req.OrgId)
	}

	// Add conversation_id to context if provided
	if req.ConversationId != "" {
		ctx = memory.WithConversationID(ctx, req.ConversationId)
	}

	// Extract JWT token from gRPC metadata and add to context
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auths := md.Get("authorization"); len(auths) > 0 {
			auth := auths[0]
			if strings.HasPrefix(auth, "Bearer ") {
				jwtToken := strings.TrimPrefix(auth, "Bearer ")
				ctx = context.WithValue(ctx, ServerJWTTokenKey, jwtToken)
			}
		}
	}

	// Add context metadata using typed keys
	for key, value := range req.Context {
		ctx = context.WithValue(ctx, grpcServerContextKey(key), value)
	}

	// Execute the agent
	result, err := s.agent.Run(ctx, req.Input)
	if err != nil {
		return &pb.RunResponse{
			Output: "",
			Error:  err.Error(),
		}, nil
	}

	return &pb.RunResponse{
		Output: result,
		Error:  "",
		Metadata: map[string]string{
			"agent_name": s.agent.GetName(),
		},
	}, nil
}

// RunStream executes the agent with streaming response
func (s *AgentServer) RunStream(req *pb.RunRequest, stream pb.AgentService_RunStreamServer) error {
	ctx := stream.Context()

	if req.Input == "" {
		return status.Error(codes.InvalidArgument, "input cannot be empty")
	}

	// Add org_id to context if provided
	if req.OrgId != "" {
		ctx = multitenancy.WithOrgID(ctx, req.OrgId)
	}

	// Add conversation_id to context if provided
	if req.ConversationId != "" {
		ctx = memory.WithConversationID(ctx, req.ConversationId)
	}

	// Extract JWT token from gRPC metadata and add to context
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auths := md.Get("authorization"); len(auths) > 0 {
			auth := auths[0]
			if strings.HasPrefix(auth, "Bearer ") {
				jwtToken := strings.TrimPrefix(auth, "Bearer ")
				ctx = context.WithValue(ctx, ServerJWTTokenKey, jwtToken)
			}
		}
	}

	// Add context metadata using typed keys
	for key, value := range req.Context {
		ctx = context.WithValue(ctx, grpcServerContextKey(key), value)
	}

	// Check if agent supports streaming
	streamingAgent, ok := interface{}(s.agent).(contracts.StreamingAgent)
	if !ok {
		// Fall back to non-streaming execution
		response, err := s.Run(ctx, req)
		if err != nil {
			return err
		}

		// Send as single chunk
		chunk := &pb.RunStreamResponse{
			Chunk:     response.Output,
			IsFinal:   true,
			EventType: pb.EventType_EVENT_TYPE_CONTENT,
			Timestamp: time.Now().UnixMilli(),
		}

		if response.Error != "" {
			chunk.Error = response.Error
			chunk.EventType = pb.EventType_EVENT_TYPE_ERROR
		}

		return stream.Send(chunk)
	}

	// Get streaming events from agent
	eventChan, err := streamingAgent.RunStream(ctx, req.Input)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to start agent streaming: %v", err)
	}

	// Stream events to client
	for event := range eventChan {
		response := &pb.RunStreamResponse{
			Chunk:     event.Content,
			EventType: s.convertEventType(event.Type),
			IsFinal:   false,
			Timestamp: event.Timestamp.UnixMilli(),
		}

		// Add metadata if present
		if event.Metadata != nil {
			response.Metadata = make(map[string]string)
			for k, v := range event.Metadata {
				if str, ok := v.(string); ok {
					response.Metadata[k] = str
				} else {
					response.Metadata[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		// Add tool call info if present
		if event.ToolCall != nil {
			response.ToolCall = &pb.ToolCall{
				Id:          event.ToolCall.ID,
				Name:        event.ToolCall.Name,
				DisplayName: event.ToolCall.DisplayName,
				Internal:    event.ToolCall.Internal,
				Arguments:   event.ToolCall.Arguments,
				Result:      event.ToolCall.Result,
				Status:      event.ToolCall.Status,
			}
		}

		// Add thinking if present
		if event.ThinkingStep != "" {
			response.Thinking = event.ThinkingStep
		}

		// Handle errors - only set error event type for non-tool errors
		// Tool errors should remain as tool result events with error status
		if event.Error != nil && event.Type != contracts.AgentEventToolResult {
			response.Error = event.Error.Error()
			response.EventType = pb.EventType_EVENT_TYPE_ERROR
		}

		// Handle completion
		if event.Type == contracts.AgentEventComplete {
			response.IsFinal = true
			response.EventType = pb.EventType_EVENT_TYPE_COMPLETE
		}

		// Send the event
		if err := stream.Send(response); err != nil {
			return status.Errorf(codes.Internal, "failed to send stream response: %v", err)
		}
	}

	// Send final completion if we haven't already
	finalResponse := &pb.RunStreamResponse{
		IsFinal:   true,
		EventType: pb.EventType_EVENT_TYPE_COMPLETE,
		Timestamp: time.Now().UnixMilli(),
	}

	return stream.Send(finalResponse)
}

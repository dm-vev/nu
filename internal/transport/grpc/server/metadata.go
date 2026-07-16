package server

import (
	"context"
	"fmt"

	pb "nu/internal/transport/grpc/pb"
)

// contextKey is a custom type for context keys to avoid collisions
type grpcServerContextKey string

// JWTTokenKey is used for JWT token context propagation - must match starops-tools exactly
const ServerJWTTokenKey grpcServerContextKey = "jwtToken"

// GetMetadata returns agent metadata
func (s *Server) GetMetadata(ctx context.Context, req *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	// Get LLM information
	llmName, llmModel := "unknown", "unknown"
	if llm := s.agent.GetLLM(); llm != nil {
		llmName = llm.Name()
		if modelGetter, ok := llm.(interface{ GetModel() string }); ok {
			llmModel = modelGetter.GetModel()
		}
		if llmModel == "" {
			llmModel = llmName
		}
	}

	// Get tool count
	tools := s.agent.GetTools()
	toolCount := len(tools)

	// Get memory info
	memoryType := "none"
	if memory := s.agent.GetMemory(); memory != nil {
		memoryType = "conversation"
	}

	return &pb.MetadataResponse{
		Name:         s.agent.GetName(),
		Description:  s.agent.GetDescription(),
		SystemPrompt: s.agent.GetSystemPrompt(), // Include system prompt for UI display
		Capabilities: []string{
			"run",
			"metadata",
			"health",
		},
		Properties: map[string]string{
			"type":       "agent",
			"version":    "1.0.0",
			"llm_name":   llmName,
			"llm_model":  llmModel,
			"tool_count": fmt.Sprintf("%d", toolCount),
			"memory":     memoryType,
		},
	}, nil
}

// GetCapabilities returns agent capabilities
func (s *Server) GetCapabilities(ctx context.Context, req *pb.CapabilitiesRequest) (*pb.CapabilitiesResponse, error) {
	// Get tool names (simplified for now)
	var toolNames []string
	// Note: This would require exposing tool information from the agent
	// For now, we'll just return basic capabilities

	var subAgentNames []string
	// Note: This would require exposing subagent information from the agent
	// For now, we'll just return basic capabilities

	return &pb.CapabilitiesResponse{
		Tools:                  toolNames,
		SubAgents:              subAgentNames,
		SupportsExecutionPlans: true, // Most agents support execution plans
		SupportsMemory:         true, // Most agents support memory
		SupportsStreaming:      true, // Now implemented!
	}, nil
}

package server

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"

	memory "nu/internal/memory/conversation"
	"nu/internal/multitenancy"
	pb "nu/internal/transport/grpc/pb"
)

// GenerateExecutionPlan generates an execution plan (if the agent supports it)
func (s *Server) GenerateExecutionPlan(ctx context.Context, req *pb.PlanRequest) (*pb.PlanResponse, error) {
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

	// Try to generate an execution plan
	plan, err := s.agent.GenerateExecutionPlan(ctx, req.Input)
	if err != nil {
		return &pb.PlanResponse{
			Error: fmt.Sprintf("Failed to generate execution plan: %v", err),
		}, nil
	}

	// Convert plan to protobuf format
	var steps []*pb.PlanStep
	for i, step := range plan.Steps {
		// Convert parameters from map[string]interface{} to map[string]string
		paramMap := make(map[string]string)
		for k, v := range step.Parameters {
			paramMap[k] = fmt.Sprintf("%v", v)
		}

		steps = append(steps, &pb.PlanStep{
			Id:          fmt.Sprintf("step_%d", i+1), // Generate ID since ExecutionStep doesn't have one
			Description: step.Description,
			ToolName:    step.ToolName,
			Parameters:  paramMap,
		})
	}

	return &pb.PlanResponse{
		PlanId:        plan.TaskID,
		FormattedPlan: formatExecutionPlan(plan),
		Steps:         steps,
	}, nil
}

// ApproveExecutionPlan approves an execution plan
func (s *Server) ApproveExecutionPlan(ctx context.Context, req *pb.ApprovalRequest) (*pb.ApprovalResponse, error) {
	// Get the plan by ID
	plan, exists := s.agent.GetTaskByID(req.PlanId)
	if !exists {
		return &pb.ApprovalResponse{
			Error: fmt.Sprintf("Plan with ID %s not found", req.PlanId),
		}, nil
	}

	var result string
	var err error

	if req.Approved {
		if req.Modifications != "" {
			// Modify the plan first
			modifiedPlan, modErr := s.agent.ModifyExecutionPlan(ctx, plan, req.Modifications)
			if modErr != nil {
				return &pb.ApprovalResponse{
					Error: fmt.Sprintf("Failed to modify plan: %v", modErr),
				}, nil
			}
			plan = modifiedPlan
		}

		// Approve and execute the plan
		result, err = s.agent.ApproveExecutionPlan(ctx, plan)
		if err != nil {
			return &pb.ApprovalResponse{
				Error: fmt.Sprintf("Failed to execute approved plan: %v", err),
			}, nil
		}
	} else {
		result = "Plan rejected by user"
	}

	return &pb.ApprovalResponse{
		Result: result,
	}, nil
}

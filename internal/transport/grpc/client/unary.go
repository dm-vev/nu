package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/metadata"

	"nu/internal/contracts"
	memory "nu/internal/memory/conversation"
	"nu/internal/multitenancy"
	pb "nu/internal/transport/grpc/pb"
)

// Run executes the remote agent with the given input
func (r *Client) Run(ctx context.Context, input string) (string, error) {
	if err := r.ensureConnected(); err != nil {
		return "", err
	}

	// Create request
	req := &pb.RunRequest{
		Input:   input,
		Context: make(map[string]string),
	}

	// Add org_id from context if available
	if orgID, _ := multitenancy.GetOrgID(ctx); orgID != "" {
		req.OrgId = orgID
	}

	// Add conversation_id from context if available
	if conversationID, ok := memory.GetConversationID(ctx); ok && conversationID != "" {
		req.ConversationId = conversationID
	}

	// Add timeout to context
	ctx, cancel := r.withTimeoutIfSet(ctx)
	defer cancel()

	// Execute with retry logic
	var lastErr error
	for attempt := 0; attempt < r.retryCount; attempt++ {
		resp, err := r.client.Run(ctx, req)
		if err != nil {
			lastErr = err
			// Exponential backoff
			if attempt < r.retryCount-1 {
				backoff := time.Duration(attempt+1) * time.Second
				time.Sleep(backoff)
			}
			continue
		}

		if resp.Error != "" {
			return "", fmt.Errorf("remote agent error: %s", resp.Error)
		}

		return resp.Output, nil
	}

	return "", fmt.Errorf("failed after %d attempts, last error: %w", r.retryCount, lastErr)
}

// RunWithAuth executes the remote agent with explicit auth token
func (r *Client) RunWithAuth(ctx context.Context, input string, authToken string) (string, error) {
	if err := r.ensureConnected(); err != nil {
		return "", err
	}

	// Create request
	req := &pb.RunRequest{
		Input:   input,
		Context: make(map[string]string),
	}

	// Add org_id from context if available
	if orgID, _ := multitenancy.GetOrgID(ctx); orgID != "" {
		req.OrgId = orgID
	}

	// Add conversation_id from context if available
	if conversationID, ok := memory.GetConversationID(ctx); ok && conversationID != "" {
		req.ConversationId = conversationID
	}

	// Add explicit auth token to gRPC metadata
	if authToken != "" {
		md := metadata.Pairs("authorization", "Bearer "+authToken)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Add timeout to context
	ctx, cancel := r.withTimeoutIfSet(ctx)
	defer cancel()

	// Execute with retry logic
	var lastErr error
	for attempt := 0; attempt < r.retryCount; attempt++ {
		resp, err := r.client.Run(ctx, req)
		if err != nil {
			lastErr = err
			// Exponential backoff
			if attempt < r.retryCount-1 {
				backoff := time.Duration(attempt+1) * time.Second
				time.Sleep(backoff)
			}
			continue
		}

		if resp.Error != "" {
			return "", fmt.Errorf("remote agent error: %s", resp.Error)
		}

		return resp.Output, nil
	}

	return "", fmt.Errorf("failed after %d attempts, last error: %w", r.retryCount, lastErr)
}

// GetMetadata retrieves metadata from the remote agent
func (r *Client) GetMetadata(ctx context.Context) (*contracts.RemoteAgentMetadata, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	metadata, err := r.client.GetMetadata(ctx, &pb.MetadataRequest{})
	if err != nil {
		return nil, err
	}
	return &contracts.RemoteAgentMetadata{
		Name:         metadata.Name,
		Description:  metadata.Description,
		SystemPrompt: metadata.SystemPrompt,
		Properties:   metadata.Properties,
	}, nil
}

// GetCapabilities retrieves capabilities from the remote agent
func (r *Client) GetCapabilities(ctx context.Context) (*pb.CapabilitiesResponse, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return r.client.GetCapabilities(ctx, &pb.CapabilitiesRequest{})
}

// Health checks the health of the remote agent service
func (r *Client) Health(ctx context.Context) (*pb.HealthResponse, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.client.Health(ctx, &pb.HealthRequest{})
}

// Ready checks if the remote agent service is ready
func (r *Client) Ready(ctx context.Context) (*pb.ReadinessResponse, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.client.Ready(ctx, &pb.ReadinessRequest{})
}

// GenerateExecutionPlan generates an execution plan via the remote agent
func (r *Client) GenerateExecutionPlan(ctx context.Context, input string) (*pb.PlanResponse, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	req := &pb.PlanRequest{
		Input:   input,
		Context: make(map[string]string),
	}

	// Add org_id from context if available
	if orgID, _ := multitenancy.GetOrgID(ctx); orgID != "" {
		req.OrgId = orgID
	}

	// Add conversation_id from context if available
	if conversationID, ok := memory.GetConversationID(ctx); ok && conversationID != "" {
		req.ConversationId = conversationID
	}

	ctx, cancel := r.withTimeoutIfSet(ctx)
	defer cancel()

	return r.client.GenerateExecutionPlan(ctx, req)
}

// ApproveExecutionPlan approves an execution plan via the remote agent
func (r *Client) ApproveExecutionPlan(ctx context.Context, planID string, approved bool, modifications string) (*pb.ApprovalResponse, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	req := &pb.ApprovalRequest{
		PlanId:        planID,
		Approved:      approved,
		Modifications: modifications,
	}

	ctx, cancel := r.withTimeoutIfSet(ctx)
	defer cancel()

	return r.client.ApproveExecutionPlan(ctx, req)
}

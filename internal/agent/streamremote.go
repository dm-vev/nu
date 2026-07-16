package agent

import (
	"context"
	"fmt"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// runRemoteStream handles streaming for remote agents
func (a *Agent) runRemoteStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	if a.remoteClient == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}

	// If orgID is set on the agent, add it to the context
	if a.orgID != "" {
		ctx = multitenancy.WithOrgID(ctx, a.orgID)
	}

	return a.remoteClient.RunStream(ctx, input)
}

// runRemoteStreamWithAuth executes a remote agent via gRPC with streaming response and explicit auth token
func (a *Agent) runRemoteStreamWithAuth(ctx context.Context, input string, authToken string) (<-chan contracts.AgentStreamEvent, error) {
	if a.remoteClient == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}

	// If orgID is set on the agent, add it to the context
	if a.orgID != "" {
		ctx = multitenancy.WithOrgID(ctx, a.orgID)
	}

	return a.remoteClient.RunStreamWithAuth(ctx, input, authToken)
}

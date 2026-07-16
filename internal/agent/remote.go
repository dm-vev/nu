package agent

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

func (a *Agent) RunWithAuth(ctx context.Context, input string, authToken string) (string, error) {
	response, err := a.runWithAuthInternal(ctx, input, authToken, false)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

func (a *Agent) RunWithAuthDetailed(ctx context.Context, input string, authToken string) (*contracts.AgentResponse, error) {
	return a.runWithAuthInternal(ctx, input, authToken, true)
}

func (a *Agent) runWithAuthInternal(ctx context.Context, input string, authToken string, detailed bool) (*contracts.AgentResponse, error) {
	startTime := time.Now()

	tracker := newUsageTracker(detailed)
	ctx = withUsageTracker(ctx, tracker)

	var response string
	var err error

	if a.isRemote {
		response, err = a.runRemoteWithAuthTracking(ctx, input, authToken)
		if err != nil {
			return nil, err
		}
	} else {
		response, err = a.runLocalWithTracking(ctx, input)
		if err != nil {
			return nil, err
		}
	}

	tracker.setExecutionTime(time.Since(startTime).Milliseconds())
	usage, execSummary, primaryModel := tracker.getResults()

	var execSum contracts.ExecutionSummary
	if execSummary != nil {
		execSum = *execSummary
	}

	return &contracts.AgentResponse{
		Content:          response,
		Usage:            usage,
		AgentName:        a.name,
		Model:            primaryModel,
		ExecutionSummary: execSum,
		Metadata: map[string]interface{}{
			"agent_name":            a.name,
			"execution_timestamp":   startTime.Unix(),
			"execution_duration_ms": time.Since(startTime).Milliseconds(),
			"auth_enabled":          true,
		},
	}, nil
}

// RunStreamWithAuth executes the agent with streaming response and explicit auth token
func (a *Agent) RunStreamWithAuth(ctx context.Context, input string, authToken string) (<-chan contracts.AgentStreamEvent, error) {
	// If this is a remote agent, delegate to remote streaming execution with auth token
	if a.isRemote {
		return a.runRemoteStreamWithAuth(ctx, input, authToken)
	}

	// For local agents, the auth token isn't used but we maintain compatibility
	return a.RunStream(ctx, input)
}

func (a *Agent) runRemoteWithTracking(ctx context.Context, input string) (string, error) {
	if a.remoteClient == nil {
		return "", fmt.Errorf("remote client not initialized")
	}

	if a.orgID != "" {
		ctx = multitenancy.WithOrgID(ctx, a.orgID)
	}

	tracker := getUsageTracker(ctx)

	if tracker != nil && tracker.detailed {
		tracker.execSummary.SubAgentCalls++

		if a.name != "" {
			found := false
			for _, used := range tracker.execSummary.UsedSubAgents {
				if used == a.name {
					found = true
					break
				}
			}
			if !found {
				tracker.execSummary.UsedSubAgents = append(tracker.execSummary.UsedSubAgents, a.name)
			}
		}
	}

	return a.remoteClient.Run(ctx, input)
}

func (a *Agent) runRemoteWithAuthTracking(ctx context.Context, input string, authToken string) (string, error) {
	if a.remoteClient == nil {
		return "", fmt.Errorf("remote client not initialized")
	}

	if a.orgID != "" {
		ctx = multitenancy.WithOrgID(ctx, a.orgID)
	}

	tracker := getUsageTracker(ctx)

	if tracker != nil && tracker.detailed {
		tracker.execSummary.SubAgentCalls++

		if a.name != "" {
			found := false
			for _, used := range tracker.execSummary.UsedSubAgents {
				if used == a.name {
					found = true
					break
				}
			}
			if !found {
				tracker.execSummary.UsedSubAgents = append(tracker.execSummary.UsedSubAgents, a.name)
			}
		}
	}

	return a.remoteClient.RunWithAuth(ctx, input, authToken)
}

// initializeRemoteAgent initializes the remote agent connection and fetches metadata
func (a *Agent) initializeRemoteAgent() error {
	// Connect to the remote agent
	// NOTE: Connection failures are non-fatal during initialization
	// This allows agents to be created even if the remote service is temporarily unavailable
	// The SDK will automatically retry connection on first actual use
	if err := a.remoteClient.Connect(); err != nil {
		// Log warning but don't fail initialization - connection will be retried on first use
		a.logger.Warn(context.Background(), fmt.Sprintf("Failed to connect to remote agent %s during initialization: %v (will retry on first use)", a.remoteURL, err), nil)
		// Return early - skip metadata fetch since connection is not available yet
		// Set default name if not provided
		if a.name == "" {
			a.name = "Remote-Agent"
		}
		return nil // Return nil to allow agent creation despite connection failure
	}

	// Fetch metadata if agent name or description is not set
	if a.name == "" || a.description == "" {
		metadata, err := a.remoteClient.GetMetadata(context.Background())
		if err != nil {
			// Don't fail if metadata fetch fails, just log and continue
			a.logger.Warn(context.Background(), fmt.Sprintf("Failed to fetch metadata from remote agent %s: %v", a.remoteURL, err), nil)
		} else {
			if a.name == "" {
				a.name = metadata.Name
			}
			if a.description == "" {
				a.description = metadata.Description
			}
		}
	}

	return nil
}

// IsRemote returns true if this is a remote agent
func (a *Agent) IsRemote() bool {
	return a.isRemote
}

// GetRemoteURL returns the URL of the remote agent (empty string if not remote)
func (a *Agent) GetRemoteURL() string {
	return a.remoteURL
}

// Disconnect closes the connection to a remote agent
func (a *Agent) Disconnect() error {
	if a.isRemote && a.remoteClient != nil {
		return a.remoteClient.Disconnect()
	}
	return nil
}

// GetRemoteMetadata returns metadata for remote agents, nil for local agents
func (a *Agent) GetRemoteMetadata() (map[string]string, error) {
	if !a.isRemote || a.remoteClient == nil {
		return nil, fmt.Errorf("not a remote agent")
	}

	metadata, err := a.remoteClient.GetMetadata(context.Background())
	if err != nil {
		return nil, err
	}

	// Convert to a simple map for easier access
	result := make(map[string]string)
	result["name"] = metadata.Name
	result["description"] = metadata.Description
	result["system_prompt"] = metadata.SystemPrompt

	// Include properties
	for k, v := range metadata.Properties {
		result[k] = v
	}

	return result, nil
}

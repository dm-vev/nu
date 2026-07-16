package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/agent/execution"
	"github.com/dm-vev/nu/agent/remote"
	"github.com/dm-vev/nu/contracts"
)

// RunWithAuth runs the agent with an explicit auth token.
func (a *Agent) RunWithAuth(ctx context.Context, input, authToken string) (string, error) {
	response, err := a.runWithAuthInternal(ctx, input, authToken, false)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// RunWithAuthDetailed returns the authenticated response and execution details.
func (a *Agent) RunWithAuthDetailed(ctx context.Context, input, authToken string) (*contracts.AgentResponse, error) {
	return a.runWithAuthInternal(ctx, input, authToken, true)
}

func (a *Agent) runWithAuthInternal(ctx context.Context, input, authToken string, detailed bool) (*contracts.AgentResponse, error) {
	startTime := time.Now()
	tracker := execution.NewTracker(detailed)
	ctx = execution.WithTracker(ctx, tracker)

	var (
		response string
		err      error
	)
	if a.isRemote {
		response, err = a.runRemoteWithAuthTracking(ctx, input, authToken)
	} else {
		response, err = a.runLocalWithTracking(ctx, input)
	}
	if err != nil {
		return nil, err
	}

	tracker.SetExecutionTime(time.Since(startTime).Milliseconds())
	usage, execSummary, primaryModel := tracker.Results()
	var summary contracts.ExecutionSummary
	if execSummary != nil {
		summary = *execSummary
	}
	return &contracts.AgentResponse{
		Content:          response,
		Usage:            usage,
		AgentName:        a.name,
		Model:            primaryModel,
		ExecutionSummary: summary,
		Metadata: map[string]interface{}{
			"agent_name":            a.name,
			"execution_timestamp":   startTime.Unix(),
			"execution_duration_ms": time.Since(startTime).Milliseconds(),
			"auth_enabled":          true,
		},
	}, nil
}

func (a *Agent) runRemoteWithTracking(ctx context.Context, input string) (string, error) {
	if a.remoteClient == nil {
		return "", fmt.Errorf("remote client not initialized")
	}
	return a.remoteService().Run(ctx, input, execution.TrackerFromContext(ctx))
}

func (a *Agent) runRemoteWithAuthTracking(ctx context.Context, input, authToken string) (string, error) {
	if a.remoteClient == nil {
		return "", fmt.Errorf("remote client not initialized")
	}
	return a.remoteService().RunWithAuth(ctx, input, authToken, execution.TrackerFromContext(ctx))
}

func (a *Agent) remoteService() remote.Service {
	return remote.Service{
		Client:      a.remoteClient,
		URL:         a.remoteURL,
		OrgID:       a.orgID,
		Name:        a.name,
		Description: a.description,
		Logger:      a.logger,
	}
}

func (a *Agent) initializeRemoteAgent() error {
	name, description := a.remoteService().Initialize()
	if a.name == "" {
		a.name = name
	}
	if a.description == "" {
		a.description = description
	}
	return nil
}

// IsRemote reports whether this is a remote agent.
func (a *Agent) IsRemote() bool { return a.isRemote }

// GetRemoteURL returns the URL of the remote agent.
func (a *Agent) GetRemoteURL() string { return a.remoteURL }

// Disconnect closes the connection to a remote agent.
func (a *Agent) Disconnect() error {
	if a.isRemote && a.remoteClient != nil {
		return a.remoteService().Disconnect()
	}
	return nil
}

// GetRemoteMetadata returns metadata for a remote agent.
func (a *Agent) GetRemoteMetadata() (map[string]string, error) {
	if !a.isRemote || a.remoteClient == nil {
		return nil, fmt.Errorf("not a remote agent")
	}
	return a.remoteService().Metadata()
}

// RunStreamWithAuth executes the agent with a streaming response and auth token.
func (a *Agent) RunStreamWithAuth(ctx context.Context, input, authToken string) (<-chan contracts.AgentStreamEvent, error) {
	if a.isRemote {
		return a.runRemoteStreamWithAuth(ctx, input, authToken)
	}
	return a.RunStream(ctx, input)
}

func (a *Agent) runRemoteStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	if a.remoteClient == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}
	return a.remoteService().Stream(ctx, input)
}

func (a *Agent) runRemoteStreamWithAuth(ctx context.Context, input, authToken string) (<-chan contracts.AgentStreamEvent, error) {
	if a.remoteClient == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}
	return a.remoteService().StreamWithAuth(ctx, input, authToken)
}

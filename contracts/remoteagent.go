package contracts

import (
	"context"
	"time"
)

// RemoteAgentMetadata is transport-neutral metadata returned by a remote agent.
type RemoteAgentMetadata struct {
	Name         string
	Description  string
	SystemPrompt string
	Properties   map[string]string
}

// RemoteAgentClient is the client behavior required by agent.Agent.
type RemoteAgentClient interface {
	Connect() error
	Disconnect() error
	SetTimeout(time.Duration)
	Run(context.Context, string) (string, error)
	RunWithAuth(context.Context, string, string) (string, error)
	RunStream(context.Context, string) (<-chan AgentStreamEvent, error)
	RunStreamWithAuth(context.Context, string, string) (<-chan AgentStreamEvent, error)
	GetMetadata(context.Context) (*RemoteAgentMetadata, error)
}

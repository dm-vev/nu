package remote

import (
	"context"

	"github.com/dm-vev/nu/agent/execution"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/telemetry"
)

// Service owns transport-neutral remote-agent lifecycle and calls.
type Service struct {
	Client      contracts.RemoteAgentClient
	URL         string
	OrgID       string
	Name        string
	Description string
	Logger      telemetry.Logger
}

func (s Service) context(ctx context.Context) context.Context {
	if s.OrgID != "" {
		return multitenancy.WithOrgID(ctx, s.OrgID)
	}
	return ctx
}

func (s Service) Run(ctx context.Context, input string, tracker *execution.Tracker) (string, error) {
	s.record(tracker)
	return s.Client.Run(s.context(ctx), input)
}

func (s Service) RunWithAuth(ctx context.Context, input, authToken string, tracker *execution.Tracker) (string, error) {
	s.record(tracker)
	return s.Client.RunWithAuth(s.context(ctx), input, authToken)
}

func (s Service) Stream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	return s.Client.RunStream(s.context(ctx), input)
}

func (s Service) StreamWithAuth(ctx context.Context, input, authToken string) (<-chan contracts.AgentStreamEvent, error) {
	return s.Client.RunStreamWithAuth(s.context(ctx), input, authToken)
}

func (s Service) record(tracker *execution.Tracker) {
	if tracker != nil {
		tracker.AddSubAgentCall(s.Name)
	}
}

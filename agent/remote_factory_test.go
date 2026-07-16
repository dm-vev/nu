package agent

import (
	"context"
	"testing"
	"time"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
)

type fakeRemoteClient struct{}

func (fakeRemoteClient) Connect() error                              { return nil }
func (fakeRemoteClient) Disconnect() error                           { return nil }
func (fakeRemoteClient) SetTimeout(time.Duration)                    {}
func (fakeRemoteClient) Run(context.Context, string) (string, error) { return "ok", nil }
func (fakeRemoteClient) RunWithAuth(context.Context, string, string) (string, error) {
	return "ok", nil
}
func (fakeRemoteClient) RunStream(context.Context, string) (<-chan contracts.AgentStreamEvent, error) {
	return make(chan contracts.AgentStreamEvent), nil
}
func (fakeRemoteClient) RunStreamWithAuth(context.Context, string, string) (<-chan contracts.AgentStreamEvent, error) {
	return make(chan contracts.AgentStreamEvent), nil
}
func (fakeRemoteClient) GetMetadata(context.Context) (*contracts.RemoteAgentMetadata, error) {
	return &contracts.RemoteAgentMetadata{Name: "remote"}, nil
}

func TestRemoteClientFactoryAppliedAfterConfig(t *testing.T) {
	a, err := NewAgent(
		WithAgentConfig(config.AgentConfig{Tools: []config.ToolConfigYAML{{Type: "agent", Name: "remote", URL: "localhost:1"}}}, nil),
		WithRemoteClientFactory(func(string) contracts.RemoteAgentClient { return fakeRemoteClient{} }),
		WithLLM(&mockLLM{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if tools := a.GetTools(); len(tools) != 1 || tools[0].Name() != "remote" {
		t.Fatalf("configured tools = %#v, want one remote tool", tools)
	}
}

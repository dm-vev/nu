package remote

import (
	"github.com/dm-vev/nu/agent"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/transport/grpc/client"
)

// NewRemoteAgent returns an agent configured with the concrete gRPC transport.
func NewRemoteAgent(url string, options ...agent.Option) (*agent.Agent, error) {
	factory := func(url string) contracts.RemoteAgentClient {
		return client.New(client.Config{URL: url})
	}
	base := []agent.Option{
		agent.WithRemoteClientFactory(factory),
		agent.WithRemoteClient(url, factory(url)),
	}
	return agent.NewAgent(append(base, options...)...)
}

// WithRemoteClients enables gRPC-backed agent tools in agent configuration.
func WithRemoteClients() agent.Option {
	return agent.WithRemoteClientFactory(func(url string) contracts.RemoteAgentClient {
		return client.New(client.Config{URL: url})
	})
}

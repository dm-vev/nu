package tools

import (
	"context"

	"github.com/dm-vev/nu/contracts"
)

// AgentToolWrapper exposes an agent as a regular tool.
type AgentToolWrapper struct {
	agent       SubAgent
	name        string
	description string
}

// Name implements contracts.Tool.Name.
func (atw *AgentToolWrapper) Name() string {
	return atw.name
}

// Description implements contracts.Tool.Description.
func (atw *AgentToolWrapper) Description() string {
	return atw.description
}

// Parameters implements contracts.Tool.Parameters.
func (atw *AgentToolWrapper) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"input": {
			Type:        "string",
			Description: "Input to send to the agent",
			Required:    true,
		},
	}
}

// Execute implements contracts.Tool.Execute.
func (atw *AgentToolWrapper) Execute(ctx context.Context, args string) (string, error) {
	return atw.agent.Run(ctx, args)
}

// Run implements contracts.Tool.Run.
func (atw *AgentToolWrapper) Run(ctx context.Context, input string) (string, error) {
	return atw.agent.Run(ctx, input)
}

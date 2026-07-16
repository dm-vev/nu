package agent

import (
	"context"
	"fmt"
	"time"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// AgentTool wraps an agent to make it callable as a tool
type AgentTool struct {
	agent       SubAgent
	name        string
	description string
	timeout     time.Duration
	logger      telemetry.Logger
	tracer      contracts.Tracer
}

// SubAgent interface defines the minimal interface needed for a sub-agent
type SubAgent interface {
	Run(ctx context.Context, input string) (string, error)
	RunStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error)
	RunDetailed(ctx context.Context, input string) (*contracts.AgentResponse, error)
	GetName() string
	GetDescription() string
}

// NewAgentTool creates a new agent tool wrapper
func NewAgentTool(agent SubAgent) *AgentTool {
	return &AgentTool{
		agent:       agent,
		name:        fmt.Sprintf("%s_agent", agent.GetName()),
		description: agent.GetDescription(),
		timeout:     30 * time.Minute,      // 30 minutes - increased timeout for long-running sub-agents
		logger:      telemetry.NewLogger(), // Default logger
	}
}

// WithTimeout sets a custom timeout for the agent tool
func (at *AgentTool) WithTimeout(timeout time.Duration) *AgentTool {
	at.timeout = timeout
	return at
}

// WithLogger sets a custom logger for the agent tool
func (at *AgentTool) WithLogger(logger telemetry.Logger) *AgentTool {
	at.logger = logger
	return at
}

// WithTracer sets a custom tracer for the agent tool
func (at *AgentTool) WithTracer(tracer contracts.Tracer) *AgentTool {
	at.tracer = tracer
	return at
}

// Name returns the name of the tool
func (at *AgentTool) Name() string {
	return at.name
}

// DisplayName implements contracts.ToolWithDisplayName.DisplayName
func (at *AgentTool) DisplayName() string {
	return fmt.Sprintf("%s Agent", at.agent.GetName())
}

// Description returns the description of what the tool does
func (at *AgentTool) Description() string {
	if at.description != "" {
		return at.description
	}
	return fmt.Sprintf("Delegate task to %s agent for specialized handling", at.agent.GetName())
}

// Internal implements contracts.InternalTool.Internal
func (at *AgentTool) Internal() bool {
	return false
}

// Parameters returns the parameters that the tool accepts
func (at *AgentTool) Parameters() map[string]contracts.ParameterSpec {
	return map[string]contracts.ParameterSpec{
		"query": {
			Type:        "string",
			Description: fmt.Sprintf("The query or task to send to the %s agent", at.agent.GetName()),
			Required:    true,
		},
		"context": {
			Type:        "object",
			Description: "Optional context information for the sub-agent",
			Required:    false,
		},
	}
}

// SetDescription allows updating the tool description
func (at *AgentTool) SetDescription(description string) {
	at.description = description
}

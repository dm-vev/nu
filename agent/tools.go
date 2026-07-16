package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/agent/tools"
	"github.com/dm-vev/nu/contracts"
)

// ToolFactory creates configured tools for an agent.
type ToolFactory = tools.Factory

// AgentToolWrapper wraps a runtime agent as a configured tool.
type AgentToolWrapper = tools.AgentToolWrapper

// NewToolFactory creates a tool factory connected to the runtime agent builder.
func NewToolFactory(remoteFactories ...func(string) contracts.RemoteAgentClient) *ToolFactory {
	factory := tools.NewFactory(remoteFactories...)
	var remoteFactory func(string) contracts.RemoteAgentClient
	if len(remoteFactories) > 0 {
		remoteFactory = remoteFactories[0]
	}
	factory.WithAgentFactory(func(url, name, description string, timeout time.Duration, requirePlanApproval *bool) (tools.SubAgent, error) {
		if remoteFactory == nil {
			return nil, fmt.Errorf("agent tool requires a remote client factory")
		}
		options := []Option{
			WithRemoteClient(url, remoteFactory(url)),
			WithRemoteTimeout(timeout),
			WithName(name),
			WithDescription(description),
		}
		if requirePlanApproval != nil {
			options = append(options, WithRequirePlanApproval(*requirePlanApproval))
		}
		return NewAgent(options...)
	})
	return factory
}

func (a *Agent) initializeConfiguredTools() {
	if len(a.configuredTools) == 0 {
		return
	}
	factory := NewToolFactory(a.remoteClientFactory)
	for _, toolConfig := range a.configuredTools {
		if toolConfig.Enabled != nil && !*toolConfig.Enabled {
			continue
		}
		tool, err := factory.CreateTool(toolConfig)
		if err != nil {
			a.logger.Warn(context.Background(), "Failed to create tool from config", map[string]interface{}{
				"tool_name": toolConfig.Name,
				"tool_type": toolConfig.Type,
				"error":     err.Error(),
			})
			continue
		}
		a.tools = deduplicateTools(append(a.tools, tool))
	}
}

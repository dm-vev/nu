package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
	agenttool "github.com/dm-vev/nu/internal/tools/agent"
	"github.com/dm-vev/nu/telemetry"
)

// WithAgents sets the sub-agents that can be called as tools
func WithAgents(subAgents ...*Agent) Option {
	return func(a *Agent) {
		a.subAgents = subAgents
		// Automatically wrap sub-agents as tools
		for _, subAgent := range subAgents {
			agentTool := agenttool.NewAgentTool(subAgent)

			// Pass logger and tracer if available on parent agent
			// Note: This will be set later in NewAgent after the agent is fully constructed
			a.tools = append(a.tools, agentTool)
		}
	}
}

// GetSubAgents returns the sub-agents slice
func (a *Agent) GetSubAgents() []*Agent {
	return a.subAgents
}

// configureSubAgentTools configures sub-agent tools with logger and tracer from parent agent
func (a *Agent) configureSubAgentTools() {
	for _, tool := range a.tools {
		// Check if this is an AgentTool by trying to cast it
		if agentTool, ok := tool.(*agenttool.AgentTool); ok {
			// Configure with parent agent's logger and tracer
			if a.tracer != nil {
				agentTool.WithTracer(a.tracer)
			}
			if a.logger != nil {
				agentTool.WithLogger(a.logger)
			}
		}
	}
}

// createSubAgentsFromConfig recursively creates sub-agents from YAML configuration
func createSubAgentsFromConfig(subAgentConfigs map[string]config.AgentConfig, variables map[string]string, llm contracts.LLM, memory contracts.Memory, tracer contracts.Tracer, logger telemetry.Logger) ([]*Agent, error) {
	if len(subAgentConfigs) == 0 {
		return nil, nil
	}

	var subAgents []*Agent
	var errors []string
	for name, config := range subAgentConfigs {
		agentOptions := []Option{WithAgentConfig(config, variables), WithName(name)}
		if llm != nil {
			agentOptions = append(agentOptions, WithLLM(llm))
		}
		if memory != nil {
			agentOptions = append(agentOptions, WithMemory(memory))
		}
		if tracer != nil {
			agentOptions = append(agentOptions, WithTracer(tracer))
		}
		if logger != nil {
			agentOptions = append(agentOptions, WithLogger(logger))
		}

		subAgent, err := NewAgent(agentOptions...)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to create sub-agent '%s': %v", name, err))
			if logger != nil {
				logger.Error(context.Background(), "Failed to create sub-agent", map[string]interface{}{
					"sub_agent_name": name,
					"error":          err.Error(),
				})
			}
			continue
		}
		subAgents = append(subAgents, subAgent)
		if logger != nil {
			logger.Info(context.Background(), "Successfully created sub-agent from YAML config", map[string]interface{}{
				"sub_agent_name":        name,
				"require_plan_approval": config.RequirePlanApproval,
				"has_sub_agents":        len(config.SubAgents) > 0,
			})
		}
	}
	if len(errors) > 0 {
		return subAgents, fmt.Errorf("some sub-agents failed to create: %s", strings.Join(errors, "; "))
	}
	return subAgents, nil
}

package agent

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/agent/plans"
	"github.com/dm-vev/nu/telemetry"
)

// NewAgent creates a new agent with the given options.
func NewAgent(options ...Option) (*Agent, error) {
	agent := &Agent{requirePlanApproval: true, maxIterations: 2}
	for _, option := range options {
		option(agent)
	}
	if agent.logger == nil {
		agent.logger = telemetry.NewLogger()
	}
	agent.initializeConfiguredTools()

	if agent.memoryConfig != nil && agent.llm != nil && agent.memory == nil {
		memoryInstance, err := CreateMemoryFromConfig(agent.memoryConfig, agent.llm)
		if err != nil {
			agent.logger.Warn(context.Background(), "Failed to create memory from config, using default", map[string]interface{}{
				"error": err.Error(),
				"type":  agent.memoryConfig["type"],
			})
		} else {
			agent.memory = memoryInstance
		}
	}
	if agent.isRemote {
		return validateRemoteAgent(agent)
	}
	return validateLocalAgent(agent)
}

func validateLocalAgent(agent *Agent) (*Agent, error) {
	if agent.llm == nil {
		return nil, fmt.Errorf("LLM is required for local agents")
	}
	if len(agent.subAgents) > 0 {
		if err := agent.validateSubAgents(); err != nil {
			return nil, fmt.Errorf("sub-agent validation failed: %w", err)
		}
		if err := validateAgentTree(agent, 5); err != nil {
			return nil, fmt.Errorf("agent tree validation failed: %w", err)
		}
	}
	agent.configureSubAgentTools()
	if err := agent.initializeMCPTools(); err != nil {
		agent.logger.Warn(context.Background(), fmt.Sprintf("Failed to initialize MCP tools: %v", err), nil)
	}
	agent.planService = plans.NewService(agent.llm, agent.getAllToolsSync(), agent.systemPrompt, agent.requirePlanApproval, agent.memory)
	return agent, nil
}

func validateRemoteAgent(agent *Agent) (*Agent, error) {
	if agent.remoteURL == "" {
		return nil, fmt.Errorf("URL is required for remote agents")
	}
	if agent.remoteClient == nil {
		return nil, fmt.Errorf("remote client is required for remote agents")
	}
	agent.remoteClient.SetTimeout(agent.remoteTimeout)
	if err := agent.initializeRemoteAgent(); err != nil {
		return nil, fmt.Errorf("failed to initialize remote agent: %w", err)
	}
	return agent, nil
}

// NewAgentWithAutoConfig creates an agent and derives task configuration from its prompt.
func NewAgentWithAutoConfig(ctx context.Context, options ...Option) (*Agent, error) {
	agent, err := NewAgent(options...)
	if err != nil {
		return nil, err
	}
	if agent.name == "" {
		agent.name = "Auto-Configured Agent"
	}
	if agent.systemPrompt == "" {
		return agent, nil
	}

	agentConfig, taskConfigs, err := config.GenerateConfigFromSystemPrompt(ctx, agent.llm, agent.systemPrompt)
	if err != nil {
		return agent, nil
	}
	taskConfigMap := make(config.TaskConfigs)
	for i, taskConfig := range taskConfigs {
		taskName := fmt.Sprintf("auto_task_%d", i+1)
		taskConfig.Agent = agent.name
		taskConfigMap[taskName] = taskConfig
	}
	agent.generatedAgentConfig = &agentConfig
	agent.generatedTaskConfigs = taskConfigMap
	return agent, nil
}

// NewAgentFromConfig creates an agent from a YAML configuration.
func NewAgentFromConfig(agentName string, configs config.AgentConfigs, variables map[string]string, options ...Option) (*Agent, error) {
	config, exists := configs[agentName]
	if !exists {
		return nil, fmt.Errorf("agent configuration for %s not found", agentName)
	}
	return NewAgent(append([]Option{WithAgentConfig(config, variables), WithName(agentName)}, options...)...)
}

// CreateAgentForTask creates an agent for a configured task.
func CreateAgentForTask(taskName string, agentConfigs config.AgentConfigs, taskConfigs config.TaskConfigs, variables map[string]string, options ...Option) (*Agent, error) {
	agentName, err := config.GetAgentForTask(taskConfigs, taskName)
	if err != nil {
		return nil, err
	}
	if taskConfig := taskConfigs[taskName]; taskConfig.ResponseFormat != nil {
		responseFormat, err := config.ConvertYAMLSchemaToResponseFormat(taskConfig.ResponseFormat)
		if err == nil && responseFormat != nil {
			options = append(options, WithResponseFormat(*responseFormat))
		}
	}
	return NewAgentFromConfig(agentName, agentConfigs, variables, options...)
}

// NewAgentFromConfigObject creates an agent from a loaded configuration.
func NewAgentFromConfigObject(ctx context.Context, config *config.AgentConfig, variables map[string]string, options ...Option) (*Agent, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	agentName := "agent"
	if config.ConfigSource != nil && config.ConfigSource.AgentName != "" {
		agentName = config.ConfigSource.AgentName
	}
	return NewAgent(append([]Option{WithAgentConfig(*config, variables), WithName(agentName)}, options...)...)
}

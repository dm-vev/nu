package agent

import (
	"context"
	"fmt"

	agentconfig "nu/internal/agent/config"
	"nu/internal/agent/plans"
	"nu/internal/telemetry"
)

// NewAgent creates a new agent with the given options
func NewAgent(options ...Option) (*Agent, error) {
	agent := &Agent{
		requirePlanApproval: true, // Default to requiring approval
		maxIterations:       2,    // Default to 2 iterations (current behavior)
	}

	for _, option := range options {
		option(agent)
	}

	// Initialize default logger if none provided
	if agent.logger == nil {
		agent.logger = telemetry.NewLogger()
	}
	agent.initializeConfiguredTools()

	// Create memory from config if specified and LLM is available
	if agent.memoryConfig != nil && agent.llm != nil && agent.memory == nil {
		memoryInstance, err := CreateMemoryFromConfig(agent.memoryConfig, agent.llm)
		if err != nil {
			// Log warning but don't fail agent creation
			if agent.logger != nil {
				agent.logger.Warn(context.Background(), "Failed to create memory from config, using default", map[string]interface{}{
					"error": err.Error(),
					"type":  agent.memoryConfig["type"],
				})
			}
		} else {
			// Apply the memory instance
			agent.memory = memoryInstance
		}
	}

	// Different validation for local vs remote agents
	if agent.isRemote {
		return validateRemoteAgent(agent)
	} else {
		return validateLocalAgent(agent)
	}
}

// validateLocalAgent validates a local agent
func validateLocalAgent(agent *Agent) (*Agent, error) {
	// Validate required fields for local agents
	if agent.llm == nil {
		return nil, fmt.Errorf("LLM is required for local agents")
	}

	// Validate sub-agents if present
	if len(agent.subAgents) > 0 {
		// Check for circular dependencies
		if err := agent.validateSubAgents(); err != nil {
			return nil, fmt.Errorf("sub-agent validation failed: %w", err)
		}

		// Validate agent tree depth (max 5 levels)
		if err := validateAgentTree(agent, 5); err != nil {
			return nil, fmt.Errorf("agent tree validation failed: %w", err)
		}
	}

	// Configure sub-agent tools with logger and tracer
	agent.configureSubAgentTools()

	// Eagerly load MCP tools during initialization to combine with manual tools
	if err := agent.initializeMCPTools(); err != nil {
		// Log warning but continue - MCP tools are optional
		agent.logger.Warn(context.Background(), fmt.Sprintf("Failed to initialize MCP tools: %v", err), nil)
	}

	// Get all tools (manual + MCP) for execution plan components
	allTools := agent.getAllToolsSync()

	// Initialize execution plan components
	agent.planStore = plans.NewExecutionPlanStore()
	agent.planGenerator = plans.NewExecutionPlanGenerator(agent.llm, allTools, agent.systemPrompt, agent.requirePlanApproval)
	agent.planExecutor = plans.NewExecutionPlanExecutor(allTools)

	return agent, nil
}

// validateRemoteAgent validates a remote agent
func validateRemoteAgent(agent *Agent) (*Agent, error) {
	// Validate required fields for remote agents
	if agent.remoteURL == "" {
		return nil, fmt.Errorf("URL is required for remote agents")
	}

	if agent.remoteClient == nil {
		return nil, fmt.Errorf("remote client is required for remote agents")
	}
	agent.remoteClient.SetTimeout(agent.remoteTimeout)

	// Test connection and fetch metadata
	if err := agent.initializeRemoteAgent(); err != nil {
		return nil, fmt.Errorf("failed to initialize remote agent: %w", err)
	}

	return agent, nil
}

// NewAgentWithAutoConfig creates a new agent with automatic configuration generation
// based on the system prompt if explicit configuration is not provided
func NewAgentWithAutoConfig(ctx context.Context, options ...Option) (*Agent, error) {
	// First create an agent with the provided options
	agent, err := NewAgent(options...)
	if err != nil {
		return nil, err
	}

	// If the agent doesn't have a name, set a default one
	if agent.name == "" {
		agent.name = "Auto-Configured Agent"
	}

	// If the system prompt is provided but no configuration was explicitly set,
	// generate configuration using the LLM
	if agent.systemPrompt != "" {
		// Generate agent and task configurations from the system prompt
		agentConfig, taskConfigs, err := agentconfig.GenerateConfigFromSystemPrompt(ctx, agent.llm, agent.systemPrompt)
		if err != nil {
			// If we fail to generate configs, just continue with the manual system prompt
			// We don't want to fail agent creation just because auto-config failed
			return agent, nil
		}

		// Create a task configuration map
		taskConfigMap := make(agentconfig.TaskConfigs)
		for i, taskConfig := range taskConfigs {
			taskName := fmt.Sprintf("auto_task_%d", i+1)
			taskConfig.Agent = agent.name // Set the task to use this agent
			taskConfigMap[taskName] = taskConfig
		}

		// Store generated configurations in agent so they can be accessed later
		agent.generatedAgentConfig = &agentConfig
		agent.generatedTaskConfigs = taskConfigMap
	}

	return agent, nil
}

// NewAgentFromConfig creates a new agent from a YAML configuration
func NewAgentFromConfig(agentName string, configs agentconfig.AgentConfigs, variables map[string]string, options ...Option) (*Agent, error) {
	config, exists := configs[agentName]
	if !exists {
		return nil, fmt.Errorf("agent configuration for %s not found", agentName)
	}

	// Add the agent config option
	configOption := WithAgentConfig(config, variables)
	nameOption := WithName(agentName)

	// Combine all options
	allOptions := append([]Option{configOption, nameOption}, options...)

	return NewAgent(allOptions...)
}

// CreateAgentForTask creates a new agent for a specific task
func CreateAgentForTask(taskName string, agentConfigs agentconfig.AgentConfigs, taskConfigs agentconfig.TaskConfigs, variables map[string]string, options ...Option) (*Agent, error) {
	agentName, err := agentconfig.GetAgentForTask(taskConfigs, taskName)
	if err != nil {
		return nil, err
	}

	// Check if task has its own response format
	taskConfig := taskConfigs[taskName]
	if taskConfig.ResponseFormat != nil {
		responseFormat, err := agentconfig.ConvertYAMLSchemaToResponseFormat(taskConfig.ResponseFormat)
		if err == nil && responseFormat != nil {
			options = append(options, WithResponseFormat(*responseFormat))
		}
	}

	return NewAgentFromConfig(agentName, agentConfigs, variables, options...)
}

// NewAgentFromConfigObject creates an agent from a pre-loaded AgentConfig object
// This is useful when you already have a loaded configuration from any source
func NewAgentFromConfigObject(ctx context.Context, config *agentconfig.AgentConfig, variables map[string]string, options ...Option) (*Agent, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create options from config
	configOption := WithAgentConfig(*config, variables)

	// Extract agent name from config source or use a default
	agentName := "agent"
	if config.ConfigSource != nil && config.ConfigSource.AgentName != "" {
		agentName = config.ConfigSource.AgentName
	}
	nameOption := WithName(agentName)

	// Combine all options
	allOptions := append([]Option{configOption, nameOption}, options...)

	return NewAgent(allOptions...)
}

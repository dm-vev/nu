package agent

import (
	"fmt"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

// GetGeneratedAgentConfig returns the automatically generated agent configuration, if any
func (a *Agent) GetGeneratedAgentConfig() *config.AgentConfig {
	return a.generatedAgentConfig
}

// GetGeneratedTaskConfigs returns the automatically generated task configurations, if any
func (a *Agent) GetGeneratedTaskConfigs() config.TaskConfigs {
	return a.generatedTaskConfigs
}

// GetName returns the agent's name
func (a *Agent) GetName() string {
	return a.name
}

// GetDescription returns the description of the agent
func (a *Agent) GetDescription() string {
	return a.description
}

// GetCapabilities returns a description of what the agent can do
func (a *Agent) GetCapabilities() string {
	if a.description != "" {
		return a.description
	}

	// If no description is set, generate one based on the system prompt
	if a.systemPrompt != "" {
		return fmt.Sprintf("Agent with system prompt: %s", a.systemPrompt)
	}

	return "A general-purpose AI agent"
}

// GetLLM returns the LLM instance (for use in custom functions)
func (a *Agent) GetLLM() contracts.LLM {
	return a.llm
}

// GetDataStore returns the datastore instance
func (a *Agent) GetDataStore() contracts.DataStore {
	return a.datastore
}

// SetDataStore sets the datastore for the agent
func (a *Agent) SetDataStore(datastore contracts.DataStore) {
	a.datastore = datastore
}

// GetTools returns the tools slice (for use in custom functions)
func (a *Agent) GetTools() []contracts.Tool {
	// Return pre-initialized tools (manual + MCP tools already combined during agent creation)
	return a.tools
}

// GetLogger returns the logger instance (for use in custom functions)
func (a *Agent) GetLogger() telemetry.Logger {
	return a.logger
}

// GetTracer returns the tracer instance (for use in custom functions)
func (a *Agent) GetTracer() contracts.Tracer {
	return a.tracer
}

// GetSystemPrompt returns the system prompt (for use in custom functions)
func (a *Agent) GetSystemPrompt() string {
	return a.systemPrompt
}

// GetConfig returns the agent's configuration for inspection
func (a *Agent) GetConfig() *config.AgentConfig {
	if a.generatedAgentConfig == nil {
		return &config.AgentConfig{}
	}
	return a.generatedAgentConfig
}

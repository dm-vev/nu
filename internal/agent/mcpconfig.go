package agent

import (
	"fmt"

	agentconfig "nu/internal/agent/config"
)

// WithMCPConfigFromJSON adds MCP servers from a JSON configuration file.
func WithMCPConfigFromJSON(filePath string) Option {
	return func(a *Agent) {
		config, err := agentconfig.LoadMCPConfigFromJSON(filePath)
		if err != nil {
			return
		}
		fmt.Printf("MCP Config loaded from JSON: %s\n", filePath)
		applyMCPConfig(a, config, nil)
	}
}

// WithMCPConfigFromYAML adds MCP servers from a YAML configuration file.
func WithMCPConfigFromYAML(filePath string) Option {
	return func(a *Agent) {
		config, err := agentconfig.LoadMCPConfigFromYAML(filePath)
		if err != nil {
			return
		}
		applyMCPConfig(a, config, nil)
	}
}

// WithMCPConfig adds MCP servers from a configuration object.
func WithMCPConfig(config *agentconfig.MCPConfiguration) Option {
	return func(a *Agent) {
		applyMCPConfig(a, config, nil)
	}
}

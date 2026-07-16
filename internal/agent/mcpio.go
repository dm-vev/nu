package agent

import (
	"strings"

	agentconfig "nu/internal/agent/config"
)

// GetMCPConfigFromAgent extracts MCP configuration from an agent.
func GetMCPConfigFromAgent(a *Agent) *agentconfig.MCPConfiguration {
	config := &agentconfig.MCPConfiguration{
		MCPServers: make(map[string]agentconfig.MCPServerConfig),
	}
	for _, lazyConfig := range a.lazyMCPConfigs {
		env := make(map[string]string)
		for _, value := range lazyConfig.Env {
			parts := strings.SplitN(value, "=", 2)
			if len(parts) == 2 {
				env[parts[0]] = parts[1]
			}
		}
		config.MCPServers[lazyConfig.Name] = agentconfig.MCPServerConfig{
			URL: lazyConfig.URL, Command: lazyConfig.Command, Args: lazyConfig.Args, Env: env,
		}
	}
	return config
}

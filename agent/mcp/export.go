package mcp

import (
	"strings"

	"github.com/dm-vev/nu/agent/config"
)

// ConfigurationFromLazy converts runtime MCP settings to public configuration.
func ConfigurationFromLazy(configs []LazyMCPConfig) *config.MCPConfiguration {
	result := &config.MCPConfiguration{MCPServers: make(map[string]config.MCPServerConfig)}
	for _, lazyConfig := range configs {
		env := make(map[string]string)
		for _, value := range lazyConfig.Env {
			parts := strings.SplitN(value, "=", 2)
			if len(parts) == 2 {
				env[parts[0]] = parts[1]
			}
		}
		result.MCPServers[lazyConfig.Name] = config.MCPServerConfig{
			URL: lazyConfig.URL, Command: lazyConfig.Command, Args: lazyConfig.Args, Env: env,
		}
	}
	return result
}

package mcp

import "github.com/dm-vev/nu/agent/config"

// LoadJSON loads an MCP configuration from a JSON file.
func LoadJSON(filePath string) (*config.MCPConfiguration, error) {
	return config.LoadMCPConfigFromJSON(filePath)
}

// LoadYAML loads an MCP configuration from a YAML file.
func LoadYAML(filePath string) (*config.MCPConfiguration, error) {
	return config.LoadMCPConfigFromYAML(filePath)
}

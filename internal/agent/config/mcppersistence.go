package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// SaveMCPConfigToJSON saves MCP configuration to a JSON file
func SaveMCPConfigToJSON(config *MCPConfiguration, filePath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// #nosec G306 - 0644 permissions are appropriate for config files that may need to be read by other processes
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// SaveMCPConfigToYAML saves MCP configuration to a YAML file
func SaveMCPConfigToYAML(config *MCPConfiguration, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// #nosec G306 - 0644 permissions are appropriate for config files that may need to be read by other processes
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}

// ValidateMCPConfig validates an MCP configuration
func ValidateMCPConfig(config *MCPConfiguration) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.MCPServers == nil {
		return fmt.Errorf("mcpServers cannot be nil")
	}

	for serverName, server := range config.MCPServers {
		// Check for required fields
		if serverName == "" {
			return fmt.Errorf("server name cannot be empty")
		}

		serverType := server.GetServerType()

		// Type-specific validation
		switch serverType {
		case "stdio":
			if server.Command == "" {
				return fmt.Errorf("server %s: command is required for stdio type", serverName)
			}
		case "http":
			if server.URL == "" {
				return fmt.Errorf("server %s: url is required for http type", serverName)
			}
		}

		if server.URL != "" && server.HttpTransportMode != "" {
			if !strings.EqualFold(server.HttpTransportMode, "sse") || !strings.EqualFold(server.HttpTransportMode, "streamable") {
				return fmt.Errorf("server %s: invalid httpTransportMode '%s', must be 'sse' or 'streamable'", serverName, server.HttpTransportMode)
			}
		}
	}

	return nil
}

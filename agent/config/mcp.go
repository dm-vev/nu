package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MCPServerConfig represents a single MCP server configuration
type MCPServerConfig struct {
	Command           string            `json:"command,omitempty" yaml:"command,omitempty"`
	Args              []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Env               map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	URL               string            `json:"url,omitempty" yaml:"url,omitempty"`
	Token             string            `json:"token,omitempty" yaml:"token,omitempty"`
	HttpTransportMode string            `json:"httpTransportMode,omitempty" yaml:"httpTransportMode,omitempty"` // "sse" or "streamable"
	AllowedTools      []string          `json:"allowedTools,omitempty" yaml:"allowedTools,omitempty"`
}

// MCPDiscoveredServerInfo represents metadata discovered from the server at runtime
type MCPDiscoveredServerInfo struct {
	Name         string                     `json:"name,omitempty"`
	Title        string                     `json:"title,omitempty"`
	Version      string                     `json:"version,omitempty"`
	Capabilities *MCPDiscoveredCapabilities `json:"capabilities,omitempty"`
}

// MCPDiscoveredCapabilities represents capabilities discovered from the server
type MCPDiscoveredCapabilities struct {
	SupportsTools     bool `json:"supportsTools,omitempty"`
	SupportsResources bool `json:"supportsResources,omitempty"`
	SupportsPrompts   bool `json:"supportsPrompts,omitempty"`
}

// GetServerType returns the server type based on configuration
func (c *MCPServerConfig) GetServerType() string {
	if c.URL != "" {
		return "http"
	}
	if c.Command != "" {
		return "stdio"
	}
	return "stdio" // Default to stdio
}

// MCPConfiguration represents the complete MCP configuration
type MCPConfiguration struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers" yaml:"mcpServers"`
	Global     *MCPGlobalConfig           `json:"global,omitempty" yaml:"global,omitempty"`
}

// MCPGlobalConfig represents global MCP settings
type MCPGlobalConfig struct {
	Timeout         string `json:"timeout,omitempty" yaml:"timeout,omitempty"` // e.g., "30s"
	RetryAttempts   int    `json:"retry_attempts,omitempty" yaml:"retry_attempts,omitempty"`
	HealthCheck     *bool  `json:"health_check,omitempty" yaml:"health_check,omitempty"`
	EnableResources *bool  `json:"enable_resources,omitempty" yaml:"enable_resources,omitempty"`
	EnablePrompts   *bool  `json:"enable_prompts,omitempty" yaml:"enable_prompts,omitempty"`
	EnableSampling  *bool  `json:"enable_sampling,omitempty" yaml:"enable_sampling,omitempty"`
	EnableSchemas   *bool  `json:"enable_schemas,omitempty" yaml:"enable_schemas,omitempty"`
	LogLevel        string `json:"log_level,omitempty" yaml:"log_level,omitempty"`
}

// LoadMCPConfigFromJSON loads MCP configuration from a JSON file
func LoadMCPConfigFromJSON(filePath string) (*MCPConfiguration, error) {
	// #nosec G304 - filePath is provided by the developer/user and expected to be a configuration file path
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var config MCPConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &config, nil
}

// LoadMCPConfigFromYAML loads MCP configuration from a YAML file
func LoadMCPConfigFromYAML(filePath string) (*MCPConfiguration, error) {
	// #nosec G304 - filePath is provided by the developer/user and expected to be a configuration file path
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var config MCPConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

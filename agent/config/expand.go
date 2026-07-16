package config

import "os"

// expandWithConfigVars expands environment variables with priority:
// 1. ConfigSource.Variables (from config server - highest priority)
// 2. OS environment variables
// 3. .env file cache
func expandWithConfigVars(s string, configVars map[string]string) string {
	if len(configVars) > 0 {
		// Use custom expansion that checks config vars first
		return os.Expand(s, func(key string) string {
			// Priority 1: Config server resolved variables
			if value, exists := configVars[key]; exists {
				return value
			}
			// Priority 2: OS environment variables
			if value, exists := os.LookupEnv(key); exists {
				return value
			}
			// Priority 3: .env cache
			if value, exists := envVarCache[key]; exists {
				return value
			}
			return ""
		})
	}
	// Fallback to standard ExpandEnv if no config source variables
	return ExpandEnv(s)
}

// ExpandWithVariables expands config-service variables before environment variables.
func ExpandWithVariables(s string, configVars map[string]string) string {
	return expandWithConfigVars(s, configVars)
}

// expandEnvironmentVariables expands environment variables in various types
func expandEnvironmentVariables(value interface{}, configVars map[string]string) interface{} {
	switch v := value.(type) {
	case string:
		return expandWithConfigVars(v, configVars)
	case map[string]interface{}:
		return expandConfigMap(v, configVars)
	case []interface{}:
		expanded := make([]interface{}, len(v))
		for i, item := range v {
			expanded[i] = expandEnvironmentVariables(item, configVars)
		}
		return expanded
	default:
		return value
	}
}

// expandConfigMap expands environment variables in a configuration map
func expandConfigMap(config map[string]interface{}, configVars map[string]string) map[string]interface{} {
	expanded := make(map[string]interface{})
	for key, value := range config {
		expanded[key] = expandEnvironmentVariables(value, configVars)
	}
	return expanded
}

// ExpandAgentConfig expands environment variables in agent configuration.
// Environment variables in the config are expanded with the following priority:
//  1. ConfigSource.Variables (from config service - highest priority)
//  2. OS environment variables
//  3. .env file cache (lowest priority)
func ExpandAgentConfig(config AgentConfig) AgentConfig {
	// Extract config variables from ConfigSource if available
	var configVars map[string]string
	if config.ConfigSource != nil && config.ConfigSource.Variables != nil {
		configVars = config.ConfigSource.Variables
	}

	expanded := config
	expanded.Role = expandWithConfigVars(config.Role, configVars)
	expanded.Goal = expandWithConfigVars(config.Goal, configVars)
	expanded.Backstory = expandWithConfigVars(config.Backstory, configVars)

	// Expand memory configuration
	if config.Memory != nil && config.Memory.Config != nil {
		expanded.Memory = &MemoryConfigYAML{
			Type:   config.Memory.Type,
			Config: expandConfigMap(config.Memory.Config, configVars),
		}
	}

	// Expand tool configurations
	if config.Tools != nil {
		expandedTools := make([]ToolConfigYAML, len(config.Tools))
		for i, tool := range config.Tools {
			expandedTools[i] = ToolConfigYAML{
				Type:        tool.Type,
				Name:        tool.Name,
				Description: expandWithConfigVars(tool.Description, configVars),
				Config:      expandConfigMap(tool.Config, configVars),
				Enabled:     tool.Enabled,
				URL:         expandWithConfigVars(tool.URL, configVars),
				Timeout:     expandWithConfigVars(tool.Timeout, configVars),
			}
		}
		expanded.Tools = expandedTools
	}

	// Expand runtime configuration
	if config.Runtime != nil {
		expanded.Runtime = &RuntimeConfigYAML{
			LogLevel:        expandWithConfigVars(config.Runtime.LogLevel, configVars),
			EnableTracing:   config.Runtime.EnableTracing,
			EnableMetrics:   config.Runtime.EnableMetrics,
			TimeoutDuration: expandWithConfigVars(config.Runtime.TimeoutDuration, configVars),
		}
	}

	// Recursively expand sub-agents configuration
	if config.SubAgents != nil {
		expandedSubAgents := make(map[string]AgentConfig)
		for name, subAgentConfig := range config.SubAgents {
			// Preserve parent's config variables for sub-agents
			if subAgentConfig.ConfigSource == nil {
				subAgentConfig.ConfigSource = &ConfigSourceMetadata{}
			}
			if subAgentConfig.ConfigSource.Variables == nil && configVars != nil {
				subAgentConfig.ConfigSource.Variables = configVars
			}
			expandedSubAgents[name] = ExpandAgentConfig(subAgentConfig) // Recursive expansion
		}
		expanded.SubAgents = expandedSubAgents
	}

	// Expand LLM provider configuration
	if config.LLMProvider != nil {
		expanded.LLMProvider = &LLMProviderYAML{
			Provider: expandWithConfigVars(config.LLMProvider.Provider, configVars),
			Model:    expandWithConfigVars(config.LLMProvider.Model, configVars),
			Config:   expandConfigMap(config.LLMProvider.Config, configVars),
		}
	}

	// Expand image generation configuration
	if config.ImageGeneration != nil {
		expanded.ImageGeneration = &ImageGenerationYAML{
			Enabled:  config.ImageGeneration.Enabled,
			Provider: expandWithConfigVars(config.ImageGeneration.Provider, configVars),
			Model:    expandWithConfigVars(config.ImageGeneration.Model, configVars),
			Config:   expandConfigMap(config.ImageGeneration.Config, configVars),
		}
		if config.ImageGeneration.Storage != nil {
			expanded.ImageGeneration.Storage = &ImageStorageYAML{
				Type: expandWithConfigVars(config.ImageGeneration.Storage.Type, configVars),
			}
			if config.ImageGeneration.Storage.Local != nil {
				expanded.ImageGeneration.Storage.Local = &LocalStorageYAML{
					Path:    expandWithConfigVars(config.ImageGeneration.Storage.Local.Path, configVars),
					BaseURL: expandWithConfigVars(config.ImageGeneration.Storage.Local.BaseURL, configVars),
				}
			}
			if config.ImageGeneration.Storage.GCS != nil {
				expanded.ImageGeneration.Storage.GCS = &GCSStorageYAML{
					Bucket:              expandWithConfigVars(config.ImageGeneration.Storage.GCS.Bucket, configVars),
					Prefix:              expandWithConfigVars(config.ImageGeneration.Storage.GCS.Prefix, configVars),
					CredentialsFile:     expandWithConfigVars(config.ImageGeneration.Storage.GCS.CredentialsFile, configVars),
					CredentialsJSON:     expandWithConfigVars(config.ImageGeneration.Storage.GCS.CredentialsJSON, configVars),
					SignedURLExpiration: expandWithConfigVars(config.ImageGeneration.Storage.GCS.SignedURLExpiration, configVars),
				}
			}
		}
		// Expand multi-turn editing configuration
		if config.ImageGeneration.MultiTurnEditing != nil {
			expanded.ImageGeneration.MultiTurnEditing = &MultiTurnEditingYAML{
				Enabled:           config.ImageGeneration.MultiTurnEditing.Enabled,
				Model:             expandWithConfigVars(config.ImageGeneration.MultiTurnEditing.Model, configVars),
				SessionTimeout:    expandWithConfigVars(config.ImageGeneration.MultiTurnEditing.SessionTimeout, configVars),
				MaxSessionsPerOrg: config.ImageGeneration.MultiTurnEditing.MaxSessionsPerOrg,
			}
		}
	}

	// Expand MCP configuration
	if config.MCP != nil {
		expandedMCP := &MCPConfiguration{
			MCPServers: make(map[string]MCPServerConfig),
			Global:     config.MCP.Global,
		}

		for serverName, serverConfig := range config.MCP.MCPServers {
			expandedServerConfig := MCPServerConfig{
				Command: expandWithConfigVars(serverConfig.Command, configVars),
				Args:    make([]string, len(serverConfig.Args)),
				Env:     make(map[string]string),
				URL:     expandWithConfigVars(serverConfig.URL, configVars),
				Token:   expandWithConfigVars(serverConfig.Token, configVars),
			}

			// Expand args
			for i, arg := range serverConfig.Args {
				expandedServerConfig.Args[i] = expandWithConfigVars(arg, configVars)
			}

			// Expand environment variables
			for key, value := range serverConfig.Env {
				expandedServerConfig.Env[key] = expandWithConfigVars(value, configVars)
			}

			expandedMCP.MCPServers[serverName] = expandedServerConfig
		}

		expanded.MCP = expandedMCP
	}

	return expanded
}

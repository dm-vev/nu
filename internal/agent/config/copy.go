package config

// deepCopyAgentConfig creates a deep copy of an AgentConfig to prevent shared state
func deepCopyAgentConfig(src *AgentConfig) *AgentConfig {
	if src == nil {
		return nil
	}

	// Create new config with basic fields (strings are immutable, safe to copy)
	dst := &AgentConfig{
		Role:      src.Role,
		Goal:      src.Goal,
		Backstory: src.Backstory,
	}

	// Deep copy pointer fields
	if src.MaxIterations != nil {
		val := *src.MaxIterations
		dst.MaxIterations = &val
	}

	if src.RequirePlanApproval != nil {
		val := *src.RequirePlanApproval
		dst.RequirePlanApproval = &val
	}

	// Deep copy ResponseFormat
	if src.ResponseFormat != nil {
		dst.ResponseFormat = &ResponseFormatConfig{
			Type:       src.ResponseFormat.Type,
			SchemaName: src.ResponseFormat.SchemaName,
		}
		// Deep copy schema definition map
		if src.ResponseFormat.SchemaDefinition != nil {
			dst.ResponseFormat.SchemaDefinition = deepCopyMap(src.ResponseFormat.SchemaDefinition)
		}
	}

	// Deep copy MCP
	if src.MCP != nil {
		dst.MCP = &MCPConfiguration{
			Global: src.MCP.Global,
		}
		// Deep copy MCPServers map
		if src.MCP.MCPServers != nil {
			dst.MCP.MCPServers = make(map[string]MCPServerConfig)
			for k, v := range src.MCP.MCPServers {
				dst.MCP.MCPServers[k] = MCPServerConfig{
					Command: v.Command,
					Args:    deepCopyStringSlice(v.Args),
					Env:     deepCopyStringMap(v.Env),
					URL:     v.URL,
					Token:   v.Token,
				}
			}
		}
	}

	// Deep copy StreamConfig
	if src.StreamConfig != nil {
		dst.StreamConfig = &StreamConfigYAML{}
		if src.StreamConfig.BufferSize != nil {
			val := *src.StreamConfig.BufferSize
			dst.StreamConfig.BufferSize = &val
		}
		if src.StreamConfig.IncludeToolProgress != nil {
			val := *src.StreamConfig.IncludeToolProgress
			dst.StreamConfig.IncludeToolProgress = &val
		}
		if src.StreamConfig.IncludeIntermediateMessages != nil {
			val := *src.StreamConfig.IncludeIntermediateMessages
			dst.StreamConfig.IncludeIntermediateMessages = &val
		}
	}

	// Deep copy LLMConfig
	if src.LLMConfig != nil {
		dst.LLMConfig = &LLMConfigYAML{}
		if src.LLMConfig.Temperature != nil {
			val := *src.LLMConfig.Temperature
			dst.LLMConfig.Temperature = &val
		}
		if src.LLMConfig.TopP != nil {
			val := *src.LLMConfig.TopP
			dst.LLMConfig.TopP = &val
		}
		if src.LLMConfig.FrequencyPenalty != nil {
			val := *src.LLMConfig.FrequencyPenalty
			dst.LLMConfig.FrequencyPenalty = &val
		}
		if src.LLMConfig.PresencePenalty != nil {
			val := *src.LLMConfig.PresencePenalty
			dst.LLMConfig.PresencePenalty = &val
		}
		if src.LLMConfig.EnableReasoning != nil {
			val := *src.LLMConfig.EnableReasoning
			dst.LLMConfig.EnableReasoning = &val
		}
		if src.LLMConfig.ReasoningBudget != nil {
			val := *src.LLMConfig.ReasoningBudget
			dst.LLMConfig.ReasoningBudget = &val
		}
		if src.LLMConfig.Reasoning != nil {
			val := *src.LLMConfig.Reasoning
			dst.LLMConfig.Reasoning = &val
		}
		dst.LLMConfig.StopSequences = deepCopyStringSlice(src.LLMConfig.StopSequences)
	}

	// Deep copy LLMProvider
	if src.LLMProvider != nil {
		dst.LLMProvider = &LLMProviderYAML{
			Provider: src.LLMProvider.Provider,
			Model:    src.LLMProvider.Model,
			Config:   deepCopyMap(src.LLMProvider.Config),
		}
	}

	// Deep copy Tools slice
	if src.Tools != nil {
		dst.Tools = make([]ToolConfigYAML, len(src.Tools))
		for i, tool := range src.Tools {
			dst.Tools[i] = ToolConfigYAML{
				Type:        tool.Type,
				Name:        tool.Name,
				Description: tool.Description,
				Config:      deepCopyMap(tool.Config),
				URL:         tool.URL,
				Timeout:     tool.Timeout,
			}
			if tool.Enabled != nil {
				val := *tool.Enabled
				dst.Tools[i].Enabled = &val
			}
		}
	}

	// Deep copy Memory
	if src.Memory != nil {
		dst.Memory = &MemoryConfigYAML{
			Type:   src.Memory.Type,
			Config: deepCopyMap(src.Memory.Config),
		}
	}

	// Deep copy Runtime
	if src.Runtime != nil {
		dst.Runtime = &RuntimeConfigYAML{
			LogLevel:        src.Runtime.LogLevel,
			TimeoutDuration: src.Runtime.TimeoutDuration,
		}
		if src.Runtime.EnableTracing != nil {
			val := *src.Runtime.EnableTracing
			dst.Runtime.EnableTracing = &val
		}
		if src.Runtime.EnableMetrics != nil {
			val := *src.Runtime.EnableMetrics
			dst.Runtime.EnableMetrics = &val
		}
	}

	// Deep copy ImageGeneration
	if src.ImageGeneration != nil {
		dst.ImageGeneration = &ImageGenerationYAML{
			Provider: src.ImageGeneration.Provider,
			Model:    src.ImageGeneration.Model,
			Config:   deepCopyMap(src.ImageGeneration.Config),
		}
		// Deep copy the Enabled bool pointer
		if src.ImageGeneration.Enabled != nil {
			val := *src.ImageGeneration.Enabled
			dst.ImageGeneration.Enabled = &val
		}
		if src.ImageGeneration.Storage != nil {
			dst.ImageGeneration.Storage = &ImageStorageYAML{
				Type: src.ImageGeneration.Storage.Type,
			}
			if src.ImageGeneration.Storage.Local != nil {
				dst.ImageGeneration.Storage.Local = &LocalStorageYAML{
					Path:    src.ImageGeneration.Storage.Local.Path,
					BaseURL: src.ImageGeneration.Storage.Local.BaseURL,
				}
			}
			if src.ImageGeneration.Storage.GCS != nil {
				dst.ImageGeneration.Storage.GCS = &GCSStorageYAML{
					Bucket:              src.ImageGeneration.Storage.GCS.Bucket,
					Prefix:              src.ImageGeneration.Storage.GCS.Prefix,
					CredentialsFile:     src.ImageGeneration.Storage.GCS.CredentialsFile,
					CredentialsJSON:     src.ImageGeneration.Storage.GCS.CredentialsJSON,
					SignedURLExpiration: src.ImageGeneration.Storage.GCS.SignedURLExpiration,
				}
			}
		}
		if src.ImageGeneration.MultiTurnEditing != nil {
			dst.ImageGeneration.MultiTurnEditing = &MultiTurnEditingYAML{
				Model:             src.ImageGeneration.MultiTurnEditing.Model,
				SessionTimeout:    src.ImageGeneration.MultiTurnEditing.SessionTimeout,
				MaxSessionsPerOrg: src.ImageGeneration.MultiTurnEditing.MaxSessionsPerOrg,
			}
			// Deep copy the Enabled bool pointer
			if src.ImageGeneration.MultiTurnEditing.Enabled != nil {
				val := *src.ImageGeneration.MultiTurnEditing.Enabled
				dst.ImageGeneration.MultiTurnEditing.Enabled = &val
			}
		}
	}

	// Deep copy SubAgents map (recursive)
	if src.SubAgents != nil {
		dst.SubAgents = make(map[string]AgentConfig)
		for name, subAgent := range src.SubAgents {
			// Recursive deep copy
			if copied := deepCopyAgentConfig(&subAgent); copied != nil {
				// Expand environment variables in sub-agent config
				expanded := ExpandAgentConfig(*copied)
				dst.SubAgents[name] = expanded
			}
		}
	}

	// Deep copy ConfigSource
	if src.ConfigSource != nil {
		dst.ConfigSource = &ConfigSourceMetadata{
			Type:        src.ConfigSource.Type,
			Source:      src.ConfigSource.Source,
			AgentID:     src.ConfigSource.AgentID,
			AgentName:   src.ConfigSource.AgentName,
			Environment: src.ConfigSource.Environment,
			LoadedAt:    src.ConfigSource.LoadedAt,
			Variables:   deepCopyStringMap(src.ConfigSource.Variables),
		}
	}

	return dst
}

// deepCopyStringSlice creates a deep copy of a string slice
func deepCopyStringSlice(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

// deepCopyStringMap creates a deep copy of a map[string]string
func deepCopyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// deepCopyMap creates a deep copy of a map[string]interface{}
func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

// deepCopyValue creates a deep copy of an interface{} value
func deepCopyValue(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case map[string]interface{}:
		return deepCopyMap(v)
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, item := range v {
			dst[i] = deepCopyValue(item)
		}
		return dst
	case []string:
		return deepCopyStringSlice(v)
	case map[string]string:
		return deepCopyStringMap(v)
	default:
		// Primitive types (string, int, bool, float64, etc.) are safe to copy by value
		return v
	}
}

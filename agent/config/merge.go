package config

import (
	"fmt"
	"os"
)

// MergeDeploymentAgentConfig merges configs with primary values overriding base values.
func MergeDeploymentAgentConfig(primary, base *AgentConfig, strategy DeploymentConfigMergeStrategy) *AgentConfig {
	if os.Getenv("DEBUG_CONFIG_MERGE") == "true" {
		debugPrintConfig(primary, "MERGE INPUT - Primary")
		debugPrintConfig(base, "MERGE INPUT - Base")
	}
	if primary == nil {
		result := deepCopyAgentConfig(base)
		if os.Getenv("DEBUG_CONFIG_MERGE") == "true" {
			debugPrintConfig(result, "MERGE OUTPUT (primary nil, returned base)")
		}
		return result
	}
	if base == nil {
		result := deepCopyAgentConfig(primary)
		if os.Getenv("DEBUG_CONFIG_MERGE") == "true" {
			debugPrintConfig(result, "MERGE OUTPUT (base nil, returned primary)")
		}
		return result
	}

	result := deepCopyAgentConfig(primary)
	mergeString := func(primaryVal, baseVal string) string {
		if primaryVal != "" {
			return primaryVal
		}
		return baseVal
	}
	result.Role = mergeString(primary.Role, base.Role)
	result.Goal = mergeString(primary.Goal, base.Goal)
	result.Backstory = mergeString(primary.Backstory, base.Backstory)

	if result.MaxIterations == nil && base.MaxIterations != nil {
		val := *base.MaxIterations
		result.MaxIterations = &val
	}
	if result.RequirePlanApproval == nil && base.RequirePlanApproval != nil {
		val := *base.RequirePlanApproval
		result.RequirePlanApproval = &val
	}
	if result.ResponseFormat == nil && base.ResponseFormat != nil {
		result.ResponseFormat = &ResponseFormatConfig{
			Type:             base.ResponseFormat.Type,
			SchemaName:       base.ResponseFormat.SchemaName,
			SchemaDefinition: deepCopyMap(base.ResponseFormat.SchemaDefinition),
		}
	}
	if result.MCP == nil && base.MCP != nil {
		result.MCP = &MCPConfiguration{Global: base.MCP.Global}
		if base.MCP.MCPServers != nil {
			result.MCP.MCPServers = make(map[string]MCPServerConfig)
			for k, v := range base.MCP.MCPServers {
				result.MCP.MCPServers[k] = MCPServerConfig{
					Command: v.Command,
					Args:    deepCopyStringSlice(v.Args),
					Env:     deepCopyStringMap(v.Env),
					URL:     v.URL,
					Token:   v.Token,
				}
			}
		}
	}
	if result.StreamConfig == nil && base.StreamConfig != nil {
		result.StreamConfig = &StreamConfigYAML{}
		if base.StreamConfig.BufferSize != nil {
			val := *base.StreamConfig.BufferSize
			result.StreamConfig.BufferSize = &val
		}
		if base.StreamConfig.IncludeToolProgress != nil {
			val := *base.StreamConfig.IncludeToolProgress
			result.StreamConfig.IncludeToolProgress = &val
		}
		if base.StreamConfig.IncludeIntermediateMessages != nil {
			val := *base.StreamConfig.IncludeIntermediateMessages
			result.StreamConfig.IncludeIntermediateMessages = &val
		}
	}
	if result.LLMConfig == nil && base.LLMConfig != nil {
		result.LLMConfig = &LLMConfigYAML{}
		if base.LLMConfig.Temperature != nil {
			val := *base.LLMConfig.Temperature
			result.LLMConfig.Temperature = &val
		}
		if base.LLMConfig.TopP != nil {
			val := *base.LLMConfig.TopP
			result.LLMConfig.TopP = &val
		}
		if base.LLMConfig.FrequencyPenalty != nil {
			val := *base.LLMConfig.FrequencyPenalty
			result.LLMConfig.FrequencyPenalty = &val
		}
		if base.LLMConfig.PresencePenalty != nil {
			val := *base.LLMConfig.PresencePenalty
			result.LLMConfig.PresencePenalty = &val
		}
		if base.LLMConfig.EnableReasoning != nil {
			val := *base.LLMConfig.EnableReasoning
			result.LLMConfig.EnableReasoning = &val
		}
		if base.LLMConfig.ReasoningBudget != nil {
			val := *base.LLMConfig.ReasoningBudget
			result.LLMConfig.ReasoningBudget = &val
		}
		if base.LLMConfig.Reasoning != nil {
			val := *base.LLMConfig.Reasoning
			result.LLMConfig.Reasoning = &val
		}
		result.LLMConfig.StopSequences = deepCopyStringSlice(base.LLMConfig.StopSequences)
	}
	if result.LLMProvider == nil && base.LLMProvider != nil {
		result.LLMProvider = &LLMProviderYAML{
			Provider: base.LLMProvider.Provider,
			Model:    base.LLMProvider.Model,
			Config:   deepCopyMap(base.LLMProvider.Config),
		}
	} else if result.LLMProvider != nil && base.LLMProvider != nil {
		merged := *result.LLMProvider
		merged.Provider = mergeString(result.LLMProvider.Provider, base.LLMProvider.Provider)
		merged.Model = mergeString(result.LLMProvider.Model, base.LLMProvider.Model)
		if merged.Config == nil && base.LLMProvider.Config != nil {
			merged.Config = deepCopyMap(base.LLMProvider.Config)
		}
		result.LLMProvider = &merged
	}

	if result.Tools == nil && base.Tools != nil {
		result.Tools = make([]ToolConfigYAML, len(base.Tools))
		for i, tool := range base.Tools {
			result.Tools[i] = ToolConfigYAML{
				Type: tool.Type, Name: tool.Name, Description: tool.Description,
				Config: deepCopyMap(tool.Config), URL: tool.URL, Timeout: tool.Timeout,
			}
			if tool.Enabled != nil {
				val := *tool.Enabled
				result.Tools[i].Enabled = &val
			}
		}
	} else if result.Tools != nil && base.Tools != nil {
		existingTools := make(map[string]bool)
		for _, tool := range result.Tools {
			existingTools[tool.Name] = true
		}
		for _, baseTool := range base.Tools {
			if !existingTools[baseTool.Name] {
				newTool := ToolConfigYAML{
					Type: baseTool.Type, Name: baseTool.Name, Description: baseTool.Description,
					Config: deepCopyMap(baseTool.Config), URL: baseTool.URL, Timeout: baseTool.Timeout,
				}
				if baseTool.Enabled != nil {
					val := *baseTool.Enabled
					newTool.Enabled = &val
				}
				result.Tools = append(result.Tools, newTool)
			}
		}
	}
	if result.Memory == nil && base.Memory != nil {
		result.Memory = &MemoryConfigYAML{Type: base.Memory.Type, Config: deepCopyMap(base.Memory.Config)}
	}
	if result.Runtime == nil && base.Runtime != nil {
		result.Runtime = &RuntimeConfigYAML{
			LogLevel: base.Runtime.LogLevel, TimeoutDuration: base.Runtime.TimeoutDuration,
		}
		if base.Runtime.EnableTracing != nil {
			val := *base.Runtime.EnableTracing
			result.Runtime.EnableTracing = &val
		}
		if base.Runtime.EnableMetrics != nil {
			val := *base.Runtime.EnableMetrics
			result.Runtime.EnableMetrics = &val
		}
	} else if result.Runtime != nil && base.Runtime != nil {
		merged := *result.Runtime
		merged.LogLevel = mergeString(result.Runtime.LogLevel, base.Runtime.LogLevel)
		merged.TimeoutDuration = mergeString(result.Runtime.TimeoutDuration, base.Runtime.TimeoutDuration)
		result.Runtime = &merged
	}
	if result.ImageGeneration == nil && base.ImageGeneration != nil {
		result.ImageGeneration = &ImageGenerationYAML{
			Enabled: base.ImageGeneration.Enabled, Provider: base.ImageGeneration.Provider,
			Model: base.ImageGeneration.Model, Config: deepCopyMap(base.ImageGeneration.Config),
		}
		if base.ImageGeneration.Storage != nil {
			result.ImageGeneration.Storage = &ImageStorageYAML{Type: base.ImageGeneration.Storage.Type}
			if base.ImageGeneration.Storage.Local != nil {
				result.ImageGeneration.Storage.Local = &LocalStorageYAML{
					Path: base.ImageGeneration.Storage.Local.Path, BaseURL: base.ImageGeneration.Storage.Local.BaseURL,
				}
			}
			if base.ImageGeneration.Storage.GCS != nil {
				result.ImageGeneration.Storage.GCS = &GCSStorageYAML{
					Bucket: base.ImageGeneration.Storage.GCS.Bucket, Prefix: base.ImageGeneration.Storage.GCS.Prefix,
					CredentialsFile:     base.ImageGeneration.Storage.GCS.CredentialsFile,
					CredentialsJSON:     base.ImageGeneration.Storage.GCS.CredentialsJSON,
					SignedURLExpiration: base.ImageGeneration.Storage.GCS.SignedURLExpiration,
				}
			}
		}
		if base.ImageGeneration.MultiTurnEditing != nil {
			result.ImageGeneration.MultiTurnEditing = &MultiTurnEditingYAML{
				Enabled: base.ImageGeneration.MultiTurnEditing.Enabled, Model: base.ImageGeneration.MultiTurnEditing.Model,
				SessionTimeout:    base.ImageGeneration.MultiTurnEditing.SessionTimeout,
				MaxSessionsPerOrg: base.ImageGeneration.MultiTurnEditing.MaxSessionsPerOrg,
			}
		}
	}

	if result.SubAgents == nil && base.SubAgents != nil {
		result.SubAgents = make(map[string]AgentConfig)
		for name, subAgent := range base.SubAgents {
			if copied := deepCopyAgentConfig(&subAgent); copied != nil {
				result.SubAgents[name] = *copied
			}
		}
	} else if result.SubAgents != nil && base.SubAgents != nil {
		for name, baseSubAgent := range base.SubAgents {
			if primarySubAgent, exists := result.SubAgents[name]; exists {
				merged := MergeDeploymentAgentConfig(&primarySubAgent, &baseSubAgent, strategy)
				result.SubAgents[name] = *merged
			} else {
				result.SubAgents[name] = baseSubAgent
			}
		}
	}
	if result.ConfigSource == nil && base.ConfigSource != nil {
		result.ConfigSource = &ConfigSourceMetadata{
			Type: base.ConfigSource.Type, Source: base.ConfigSource.Source, AgentID: base.ConfigSource.AgentID,
			AgentName: base.ConfigSource.AgentName, Environment: base.ConfigSource.Environment,
			LoadedAt: base.ConfigSource.LoadedAt, Variables: deepCopyStringMap(base.ConfigSource.Variables),
		}
	} else if result.ConfigSource != nil && base.ConfigSource != nil {
		result.ConfigSource.Type = string(DeploymentConfigSourceMerged)
		result.ConfigSource.Source = fmt.Sprintf("merged(%s + %s)", result.ConfigSource.Source, base.ConfigSource.Source)
		if result.ConfigSource.Variables == nil && base.ConfigSource.Variables != nil {
			result.ConfigSource.Variables = deepCopyStringMap(base.ConfigSource.Variables)
		} else if result.ConfigSource.Variables != nil && base.ConfigSource.Variables != nil {
			merged := make(map[string]string)
			for k, v := range base.ConfigSource.Variables {
				merged[k] = v
			}
			for k, v := range result.ConfigSource.Variables {
				merged[k] = v
			}
			result.ConfigSource.Variables = merged
		}
	}
	if os.Getenv("DEBUG_CONFIG_MERGE") == "true" {
		debugPrintConfig(result, "MERGE OUTPUT (final merged config)")
	}
	return result
}

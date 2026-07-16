package config

import "nu/internal/contracts"

// ConvertYAMLSchemaToResponseFormat converts a ResponseFormatConfig to contracts.ResponseFormat
func ConvertYAMLSchemaToResponseFormat(config *ResponseFormatConfig) (*contracts.ResponseFormat, error) {
	if config == nil {
		return nil, nil
	}

	schema := contracts.JSONSchema(config.SchemaDefinition)
	return &contracts.ResponseFormat{
		Type:   contracts.ResponseFormatType(config.Type),
		Name:   config.SchemaName,
		Schema: schema,
	}, nil
}

// convertStreamConfigYAMLToInterface converts StreamConfigYAML to contracts.StreamConfig
func ConvertStreamConfig(config *StreamConfigYAML) *contracts.StreamConfig {
	if config == nil {
		return nil
	}

	streamConfig := &contracts.StreamConfig{}

	if config.BufferSize != nil {
		streamConfig.BufferSize = *config.BufferSize
	}
	if config.IncludeToolProgress != nil {
		streamConfig.IncludeToolProgress = *config.IncludeToolProgress
	}
	if config.IncludeIntermediateMessages != nil {
		streamConfig.IncludeIntermediateMessages = *config.IncludeIntermediateMessages
	}

	return streamConfig
}

// convertCacheConfigYAMLToInterface converts CacheConfigYAML to contracts.CacheConfig
func ConvertCacheConfig(cfg *CacheConfigYAML) *contracts.CacheConfig {
	if cfg == nil {
		return nil
	}

	cc := &contracts.CacheConfig{}
	if cfg.CacheSystemMessage != nil {
		cc.CacheSystemMessage = *cfg.CacheSystemMessage
	}
	if cfg.CacheTools != nil {
		cc.CacheTools = *cfg.CacheTools
	}
	if cfg.CacheConversation != nil {
		cc.CacheConversation = *cfg.CacheConversation
	}
	if cfg.CacheTTL != nil {
		cc.CacheTTL = *cfg.CacheTTL
	}
	return cc
}

// convertLLMConfigYAMLToInterface converts LLMConfigYAML to contracts.LLMConfig
func ConvertLLMConfig(config *LLMConfigYAML) *contracts.LLMConfig {
	if config == nil {
		return nil
	}

	llmConfig := &contracts.LLMConfig{}

	if config.Temperature != nil {
		llmConfig.Temperature = *config.Temperature
	}
	if config.TopP != nil {
		llmConfig.TopP = *config.TopP
	}
	if config.FrequencyPenalty != nil {
		llmConfig.FrequencyPenalty = *config.FrequencyPenalty
	}
	if config.PresencePenalty != nil {
		llmConfig.PresencePenalty = *config.PresencePenalty
	}
	if config.StopSequences != nil {
		llmConfig.StopSequences = config.StopSequences
	}
	if config.EnableReasoning != nil {
		llmConfig.EnableReasoning = *config.EnableReasoning
	}
	if config.ReasoningBudget != nil {
		llmConfig.ReasoningBudget = *config.ReasoningBudget
	}
	if config.Reasoning != nil {
		llmConfig.Reasoning = *config.Reasoning
	}

	return llmConfig
}

// convertMemoryConfigYAMLToInterface converts MemoryConfigYAML to runtime memory config (placeholder)
// Note: This returns config data that will be used at runtime to create actual memory instances
func ConvertMemoryConfig(config *MemoryConfigYAML) map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"type": config.Type,
	}

	if config.Config != nil {
		for k, v := range config.Config {
			result[k] = v
		}
	}

	return result
}

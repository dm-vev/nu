package anthropic

import "nu/internal/contracts"

func WithReasoning(reasoning string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.Reasoning = reasoning
	}
}

func WithCacheSystemMessage() contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.CacheConfig == nil {
			options.CacheConfig = &contracts.CacheConfig{}
		}
		options.CacheConfig.CacheSystemMessage = true
	}
}

func WithCacheTools() contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.CacheConfig == nil {
			options.CacheConfig = &contracts.CacheConfig{}
		}
		options.CacheConfig.CacheTools = true
	}
}

func WithCacheConversation() contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.CacheConfig == nil {
			options.CacheConfig = &contracts.CacheConfig{}
		}
		options.CacheConfig.CacheConversation = true
	}
}

func WithCacheTTL(ttl string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.CacheConfig == nil {
			options.CacheConfig = &contracts.CacheConfig{}
		}
		options.CacheConfig.CacheTTL = ttl
	}
}

func WithTemperature(temperature float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.LLMConfig.Temperature = temperature }
}

func WithTopP(topP float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.LLMConfig.TopP = topP }
}

func WithFrequencyPenalty(frequencyPenalty float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.LLMConfig.FrequencyPenalty = frequencyPenalty }
}

func WithPresencePenalty(presencePenalty float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.LLMConfig.PresencePenalty = presencePenalty }
}

func WithStopSequences(stopSequences []string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.LLMConfig.StopSequences = stopSequences }
}

func WithSystemMessage(systemMessage string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.SystemMessage = systemMessage }
}

func WithResponseFormat(format contracts.ResponseFormat) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.ResponseFormat = &format }
}

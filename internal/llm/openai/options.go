package openai

import (
	"nu/internal/contracts"
	"nu/internal/llm"
	"nu/internal/telemetry"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// OpenAIOption represents an option for configuring the OpenAI client.
type Option func(*Client)

// WithOpenAIModel sets the model for the OpenAI client.
func WithModel(model string) Option {
	return func(c *Client) {
		c.Model = model
	}
}

// WithOpenAILogger sets the logger for the OpenAI client.
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithOpenAIRetry configures retry policy for the client.
func WithRetry(opts ...llm.RetryOption) Option {
	return func(c *Client) {
		c.retryExecutor = llm.NewRetryExecutor(llm.NewRetryPolicy(opts...))
	}
}

// WithOpenAIBaseURL sets the base URL for the OpenAI client.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
		c.Client = openai.NewClient(option.WithAPIKey(c.apiKey), option.WithBaseURL(baseURL))
		c.ChatService = openai.NewChatService(option.WithAPIKey(c.apiKey), option.WithBaseURL(baseURL))
		c.ResponseService = openai.NewClient(option.WithAPIKey(c.apiKey), option.WithBaseURL(baseURL))
	}
}

// WithOpenAITemperature creates a GenerateOption to set the temperature.
func WithTemperature(temperature float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.LLMConfig.Temperature = temperature
	}
}

// WithOpenAITopP creates a GenerateOption to set top_p.
func WithTopP(topP float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.LLMConfig.TopP = topP
	}
}

// WithOpenAIFrequencyPenalty creates a GenerateOption to set the frequency penalty.
func WithFrequencyPenalty(frequencyPenalty float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.LLMConfig.FrequencyPenalty = frequencyPenalty
	}
}

// WithOpenAIPresencePenalty creates a GenerateOption to set the presence penalty.
func WithPresencePenalty(presencePenalty float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.LLMConfig.PresencePenalty = presencePenalty
	}
}

// WithOpenAIStopSequences creates a GenerateOption to set the stop sequences.
func WithStopSequences(stopSequences []string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.LLMConfig.StopSequences = stopSequences
	}
}

// WithOpenAISystemMessage creates a GenerateOption to set the system message.
func WithSystemMessage(systemMessage string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.SystemMessage = systemMessage
	}
}

// WithOpenAIResponseFormat creates a GenerateOption to set the response format.
func WithResponseFormat(format contracts.ResponseFormat) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.ResponseFormat = &format
	}
}

// WithOpenAIReasoning sets the reasoning effort for OpenAI reasoning models.
func WithReasoning(reasoning string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.Reasoning = reasoning
	}
}

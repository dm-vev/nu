package azureopenai

import (
	"context"
	"strings"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm"
	"github.com/dm-vev/nu/telemetry"
)

// AzureOpenAIOption represents an option for configuring the Azure OpenAI client
type Option func(*Client)

// WithAzureOpenAIModel sets the model for the Azure OpenAI client
func WithModel(model string) Option {
	return func(c *Client) { c.Model = model }
}

// WithAzureOpenAIDeployment sets the deployment name for the Azure OpenAI client
func WithDeployment(deployment string) Option {
	return func(c *Client) { c.deployment = deployment }
}

// WithAzureOpenAIAPIVersion sets the API version for the Azure OpenAI client
func WithAPIVersion(apiVersion string) Option {
	return func(c *Client) { c.apiVersion = apiVersion }
}

// WithAzureOpenAIRegion sets the Azure region for the Azure OpenAI client
func WithRegion(region string) Option {
	return func(c *Client) { c.region = region }
}

// WithAzureOpenAIResourceName sets the Azure resource name for the Azure OpenAI client
func WithResourceName(resourceName string) Option {
	return func(c *Client) { c.resourceName = resourceName }
}

// WithAzureOpenAILogger sets the logger for the Azure OpenAI client
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) { c.logger = logger }
}

// WithAzureOpenAIRetry configures retry policy for the client
func WithRetry(opts ...llm.RetryOption) Option {
	return func(c *Client) {
		c.retryExecutor = llm.NewRetryExecutor(llm.NewRetryPolicy(opts...))
	}
}

// WithAzureOpenAIBaseURL sets the base URL for the Azure OpenAI client
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
		// Recreate the client and services with the new base URL
		c.recreateClients()
	}
}

// azureOpenAIIsReasoningModel returns true if the model is a reasoning model that requires temperature = 1
func azureOpenAIIsReasoningModel(model string) bool {
	reasoningModels := []string{
		"o1-", "o1-mini", "o1-preview",
		"o3-", "o3-mini",
		"o4-", "o4-mini",
		"gpt-5", "gpt-5-mini", "gpt-5-nano",
	}

	for _, prefix := range reasoningModels {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

// getTemperatureForModel returns the appropriate temperature for a model
func (c *Client) getTemperatureForModel(requestedTemp float64) float64 {
	if azureOpenAIIsReasoningModel(c.Model) {
		if requestedTemp != 1.0 {
			c.logger.Debug(context.Background(), "Overriding temperature for reasoning model", map[string]interface{}{
				"model":                 c.Model,
				"requested_temperature": requestedTemp,
				"forced_temperature":    1.0,
				"reason":                "reasoning models only support temperature = 1",
			})
		}
		return 1.0
	}
	return requestedTemp
}

// WithAzureOpenAITemperature creates a GenerateOption to set the temperature
func WithTemperature(temperature float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.Temperature = temperature
	}
}

// WithAzureOpenAITopP creates a GenerateOption to set the top_p
func WithTopP(topP float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.TopP = topP
	}
}

// WithAzureOpenAIFrequencyPenalty creates a GenerateOption to set the frequency penalty
func WithFrequencyPenalty(frequencyPenalty float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.FrequencyPenalty = frequencyPenalty
	}
}

// WithAzureOpenAIPresencePenalty creates a GenerateOption to set the presence penalty
func WithPresencePenalty(presencePenalty float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.PresencePenalty = presencePenalty
	}
}

// WithAzureOpenAIStopSequences creates a GenerateOption to set the stop sequences
func WithStopSequences(stopSequences []string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.StopSequences = stopSequences
	}
}

// WithAzureOpenAISystemMessage creates a GenerateOption to set the system message
func WithSystemMessage(systemMessage string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.SystemMessage = systemMessage }
}

// WithAzureOpenAIResponseFormat creates a GenerateOption to set the response format
func WithResponseFormat(format contracts.ResponseFormat) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) { options.ResponseFormat = &format }
}

// WithAzureOpenAIReasoning creates a GenerateOption to set the reasoning effort for reasoning models
// For OpenAI reasoning models (o1, o3, o4, gpt-5 series), valid values are:
// "minimal", "low", "medium", "high"
// This parameter is only used with reasoning models and is ignored for standard models.
func WithReasoning(reasoning string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.Reasoning = reasoning
	}
}

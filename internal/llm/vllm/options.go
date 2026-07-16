package vllm

import (
	"net/http"

	"nu/internal/contracts"
	"nu/internal/llm"
	"nu/internal/telemetry"
)

// VLLMOption represents an option for configuring the vLLM client
type Option func(*Client)

// WithVLLMModel sets the model for the vLLM client
func WithModel(model string) Option {
	return func(c *Client) {
		c.Model = model
	}
}

// WithVLLMLogger sets the logger for the vLLM client
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithVLLMRetry configures retry policy for the client
func WithRetry(opts ...llm.RetryOption) Option {
	return func(c *Client) {
		c.retryExecutor = llm.NewRetryExecutor(llm.NewRetryPolicy(opts...))
	}
}

// WithVLLMBaseURL sets the base URL for the vLLM API
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithVLLMHTTPClient sets the HTTP client for the vLLM client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// GenerateOption functions for vLLM
func WithTemperature(temperature float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.Temperature = temperature
	}
}

func WithTopP(topP float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.TopP = topP
	}
}

func WithStopSequences(stopSequences []string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.StopSequences = stopSequences
	}
}

func WithSystemMessage(systemMessage string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.SystemMessage = systemMessage
	}
}

func WithResponseFormat(format contracts.ResponseFormat) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.ResponseFormat = &format
	}
}

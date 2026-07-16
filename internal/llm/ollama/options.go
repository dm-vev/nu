package ollama

import (
	"net/http"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm"
	"github.com/dm-vev/nu/telemetry"
)

// OllamaOption represents an option for configuring the Ollama client
type Option func(*Client)

// WithOllamaModel sets the model for the Ollama client
func WithModel(model string) Option {
	return func(c *Client) {
		c.Model = model
	}
}

// WithOllamaLogger sets the logger for the Ollama client
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithOllamaRetry configures retry policy for the client
func WithRetry(opts ...llm.RetryOption) Option {
	return func(c *Client) {
		c.retryExecutor = llm.NewRetryExecutor(llm.NewRetryPolicy(opts...))
	}
}

// WithOllamaBaseURL sets the base URL for the Ollama API
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithOllamaHTTPClient sets the HTTP client for the Ollama client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// GenerateOption functions for Ollama
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

package gemini

import (
	"context"

	"google.golang.org/genai"

	"nu/internal/contracts"
	"nu/internal/llm"
	"nu/internal/telemetry"
)

// GeminiOption represents an option for configuring the Gemini client
type Option func(*Client)

// WithGeminiModel sets the model for the Gemini client
func WithModel(model string) Option {
	return func(c *Client) {
		c.model = model
	}
}

// WithGeminiLogger sets the logger for the Gemini client
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithGeminiRetry configures retry policy for the client
func WithRetry(opts ...llm.RetryOption) Option {
	return func(c *Client) {
		c.retryExecutor = llm.NewRetryExecutor(llm.NewRetryPolicy(opts...))
	}
}

// WithGeminiAPIKey sets the API key for Gemini API backend
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithGeminiBaseURL sets the base URL for the Gemini client (not used with genai package)
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		// Note: baseURL is not used with the genai package as it manages the endpoint internally
		c.logger.Warn(context.Background(), "BaseURL option is not supported with Gemini API client", nil)
	}
}

// WithGeminiClient injects an already initialized genai.Client. If set, NewGemini won't build a new client
func WithClient(existing *genai.Client) Option {
	return func(c *Client) {
		c.genaiClient = existing
	}
}

// WithGeminiBackend sets the backend for the Gemini client
func WithBackend(backend genai.Backend) Option {
	return func(c *Client) {
		c.backend = backend
	}
}

// WithGeminiProjectID sets the GCP project ID for Vertex AI backend
func WithProjectID(projectID string) Option {
	return func(c *Client) {
		c.projectID = projectID
	}
}

// WithGeminiLocation sets the GCP location for Vertex AI backend
func WithLocation(location string) Option {
	return func(c *Client) {
		c.location = location
	}
}

// WithGeminiCredentialsFile sets the path to a service account key file for Vertex AI authentication.
// If both WithGeminiCredentialsFile and WithGeminiCredentialsJSON are provided, JSON credentials take precedence.
// The file should contain a valid Google Cloud service account key in JSON format.
func WithCredentialsFile(credentialsFile string) Option {
	return func(c *Client) {
		c.credentialsFile = credentialsFile
	}
}

// WithGeminiCredentialsJSON sets the service account key JSON bytes for Vertex AI authentication.
// If both WithGeminiCredentialsFile and WithGeminiCredentialsJSON are provided, JSON credentials take precedence.
// The bytes should contain a valid Google Cloud service account key in JSON format.
func WithCredentialsJSON(credentialsJSON []byte) Option {
	return func(c *Client) {
		c.credentialsJSON = credentialsJSON
	}
}

// WithGeminiMaxOutputTokens sets the maximum number of output tokens to generate.
// This limits the length of the model's response.
func WithMaxOutputTokens(maxTokens int32) Option {
	return func(c *Client) {
		c.maxOutputTokens = &maxTokens
	}
}

// WithGeminiTemperature creates a GenerateOption to set the temperature
func WithTemperature(temperature float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.Temperature = temperature
	}
}

// WithGeminiTopP creates a GenerateOption to set the top_p
func WithTopP(topP float64) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.TopP = topP
	}
}

// WithGeminiStopSequences creates a GenerateOption to set the stop sequences
func WithStopSequences(stopSequences []string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.StopSequences = stopSequences
	}
}

// WithGeminiSystemMessage creates a GenerateOption to set the system message
func WithSystemMessage(systemMessage string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.SystemMessage = systemMessage
	}
}

// WithGeminiResponseFormat creates a GenerateOption to set the response format
func WithResponseFormat(format contracts.ResponseFormat) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		options.ResponseFormat = &format
	}
}

// WithGeminiReasoning creates a GenerateOption to set the reasoning mode
// reasoning can be "none" (direct answers), "minimal" (brief explanations),
// or "comprehensive" (detailed step-by-step reasoning)
func WithReasoning(reasoning string) contracts.GenerateOption {
	return func(options *contracts.GenerateOptions) {
		if options.LLMConfig == nil {
			options.LLMConfig = &contracts.LLMConfig{}
		}
		options.LLMConfig.Reasoning = reasoning
	}
}

// Thinking-related client options (for configuring the GeminiClient)

// WithGeminiThinking creates a client GeminiOption to enable/disable thinking
func WithThinking(enabled bool) Option {
	return func(c *Client) {
		if c.thinkingConfig == nil {
			defaultConfig := DefaultThinkingConfig()
			c.thinkingConfig = &defaultConfig
		}
		c.thinkingConfig.IncludeThoughts = enabled
	}
}

// WithGeminiThinkingBudget creates a client GeminiOption to set thinking token budget
func WithThinkingBudget(budget int32) Option {
	return func(c *Client) {
		if c.thinkingConfig == nil {
			defaultConfig := DefaultThinkingConfig()
			c.thinkingConfig = &defaultConfig
		}
		c.thinkingConfig.ThinkingBudget = &budget
	}
}

// WithGeminiDynamicThinking creates a client GeminiOption to enable dynamic thinking (no fixed budget)
func WithDynamicThinking() Option {
	return func(c *Client) {
		if c.thinkingConfig == nil {
			defaultConfig := DefaultThinkingConfig()
			c.thinkingConfig = &defaultConfig
		}
		c.thinkingConfig.ThinkingBudget = nil // nil means dynamic
		c.thinkingConfig.IncludeThoughts = true
	}
}

// WithGeminiThoughtSignatures creates a client GeminiOption to set thought signatures for multi-turn context
func WithThoughtSignatures(signatures [][]byte) Option {
	return func(c *Client) {
		if c.thinkingConfig == nil {
			defaultConfig := DefaultThinkingConfig()
			c.thinkingConfig = &defaultConfig
		}
		c.thinkingConfig.ThoughtSignatures = signatures
	}
}

// WithGeminiThinkingConfig creates a client GeminiOption to set complete thinking configuration
func WithThinkingConfig(config ThinkingConfig) Option {
	return func(c *Client) {
		c.thinkingConfig = &config
	}
}

package deepseek

import (
	"net/http"
	"time"

	"nu/internal/llm"
	"nu/internal/telemetry"
)

const (
	// DefaultBaseURL is the default DeepSeek API base URL
	DefaultBaseURL = "https://api.deepseek.com"

	// DeepSeekDefaultModel is the default DeepSeek model
	DefaultModel = "deepseek-chat"

	// DefaultMaxIterations is the default maximum number of tool calling iterations
	DefaultMaxIterations = 10
)

// DeepSeekClient implements the LLM interface for DeepSeek
type Client struct {
	APIKey        string
	Model         string
	BaseURL       string
	HTTPClient    *http.Client
	logger        telemetry.Logger
	retryExecutor *llm.RetryExecutor
}

// NewDeepSeek creates a new DeepSeek client
func NewClient(apiKey string, options ...Option) *Client {
	client := &Client{
		APIKey:  apiKey,
		Model:   DefaultModel,
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: telemetry.NewLogger(),
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client
}

// Name returns the name of the LLM provider
func (c *Client) Name() string {
	return "deepseek"
}

// SupportsStreaming returns true if this LLM supports streaming
func (c *Client) SupportsStreaming() bool {
	return true
}

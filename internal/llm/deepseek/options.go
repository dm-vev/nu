package deepseek

import (
	"net/http"
	"strings"

	"github.com/dm-vev/nu/internal/llm"
	"github.com/dm-vev/nu/telemetry"
)

// DeepSeekOption represents an option for configuring the DeepSeek client
type Option func(*Client)

// WithDeepSeekModel sets the model for the DeepSeek client
func WithModel(model string) Option {
	return func(c *Client) {
		c.Model = model
	}
}

// WithDeepSeekLogger sets the logger for the DeepSeek client
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithDeepSeekRetry configures retry policy for the client
func WithRetry(opts ...llm.RetryOption) Option {
	return func(c *Client) {
		c.retryExecutor = llm.NewRetryExecutor(llm.NewRetryPolicy(opts...))
	}
}

// WithDeepSeekBaseURL sets the base URL for the DeepSeek client
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithDeepSeekHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = client
	}
}

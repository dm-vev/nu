package ollama

import (
	"net/http"
	"time"

	"github.com/dm-vev/nu/internal/llm"
	"github.com/dm-vev/nu/telemetry"
)

// OllamaClient implements the LLM interface for Ollama
type Client struct {
	BaseURL       string
	HTTPClient    *http.Client
	Model         string
	logger        telemetry.Logger
	retryExecutor *llm.RetryExecutor
}

// NewOllama creates a new Ollama client
func NewClient(options ...Option) *Client {
	// Create client with default options
	client := &Client{
		BaseURL:    "http://localhost:11434",
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Model:      "qwen3:0.6b",
		logger:     telemetry.NewLogger(),
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client
}

// Name returns the name of the LLM provider
func (c *Client) Name() string {
	return "ollama"
}

// SupportsStreaming returns false as streaming is not yet implemented for Ollama
func (c *Client) SupportsStreaming() bool {
	return false
}

// GetModel returns the model name being used
func (c *Client) GetModel() string {
	return c.Model
}

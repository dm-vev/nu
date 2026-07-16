package vllm

import (
	"net/http"
	"time"

	"nu/internal/llm"
	"nu/internal/telemetry"
)

// VLLMClient implements the LLM interface for vLLM
type Client struct {
	BaseURL       string
	HTTPClient    *http.Client
	Model         string
	logger        telemetry.Logger
	retryExecutor *llm.RetryExecutor
}

// NewVLLM creates a new vLLM client
func NewClient(options ...Option) *Client {
	// Create client with default options
	client := &Client{
		BaseURL:    "http://localhost:8000",
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Model:      "llama-2-7b",
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
	return "vllm"
}

// SupportsStreaming returns false as streaming is not yet implemented for VLLM
func (c *Client) SupportsStreaming() bool {
	return false
}

// GetModel returns the model name being used
func (c *Client) GetModel() string {
	return c.Model
}

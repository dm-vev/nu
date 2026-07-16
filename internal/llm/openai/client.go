package openai

import (
	"nu/internal/llm"
	"nu/internal/telemetry"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type openAIContextKey string

const openAIOrganizationKey openAIContextKey = "organization"

// OpenAIClient implements the LLM interface for OpenAI.
type Client struct {
	Client          openai.Client
	ChatService     openai.ChatService
	ResponseService openai.Client
	Model           string
	apiKey          string
	baseURL         string
	logger          telemetry.Logger
	retryExecutor   *llm.RetryExecutor
}

// NewOpenAI creates a new OpenAI client.
func NewClient(apiKey string, options ...Option) *Client {
	client := &Client{
		Client:          openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL("https://api.openai.com/v1")),
		ChatService:     openai.NewChatService(option.WithAPIKey(apiKey), option.WithBaseURL("https://api.openai.com/v1")),
		ResponseService: openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL("https://api.openai.com/v1")),
		Model:           "gpt-4o-mini",
		apiKey:          apiKey,
		baseURL:         "https://api.openai.com/v1",
		logger:          telemetry.NewLogger(),
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// Name implements contracts.LLM.Name.
func (c *Client) Name() string {
	return "openai"
}

// SupportsStreaming implements contracts.LLM.SupportsStreaming.
func (c *Client) SupportsStreaming() bool {
	return true
}

// GetModel returns the model name being used.
func (c *Client) GetModel() string {
	return c.Model
}

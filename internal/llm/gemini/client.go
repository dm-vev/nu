package gemini

import (
	"context"

	"google.golang.org/genai"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm"
	"github.com/dm-vev/nu/telemetry"
)

// GeminiClient implements the LLM interface for Google Gemini API
type Client struct {
	genaiClient     *genai.Client
	apiKey          string
	model           string
	backend         genai.Backend
	projectID       string
	location        string
	credentialsFile string
	credentialsJSON []byte
	logger          telemetry.Logger
	retryExecutor   *llm.RetryExecutor
	thinkingConfig  *ThinkingConfig
	maxOutputTokens *int32 // Maximum number of output tokens to generate
}

// applyMaxOutputTokens applies the client's max output tokens to the generation config if set
func (c *Client) applyMaxOutputTokens(genConfig **genai.GenerationConfig) {
	if c.maxOutputTokens != nil {
		if *genConfig == nil {
			*genConfig = &genai.GenerationConfig{}
		}
		// MaxOutputTokens expects int32 value, not pointer
		maxTokens := *c.maxOutputTokens
		(*genConfig).MaxOutputTokens = maxTokens
	}
}

// Name implements contracts.LLM.Name
func (c *Client) Name() string {
	return "gemini"
}

// SupportsStreaming implements contracts.LLM.SupportsStreaming
func (c *Client) SupportsStreaming() bool {
	return true
}

// GetModel returns the model name being used
func (c *Client) GetModel() string {
	return c.model
}

// buildContentsWithMemory builds Gemini contents from memory messages and current prompt
func (c *Client) buildContentsWithMemory(ctx context.Context, prompt string, params *contracts.GenerateOptions) []*genai.Content {
	builder := geminiNewMessageHistoryBuilder(c.logger)
	return builder.buildContents(ctx, prompt, params)
}

// GenerateDetailed generates text and returns detailed response information including token usage
func (c *Client) GenerateDetailed(ctx context.Context, prompt string, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	return c.generateInternal(ctx, prompt, options...)
}

// GenerateWithToolsDetailed generates text with tools and returns detailed response information including token usage
func (c *Client) GenerateWithToolsDetailed(ctx context.Context, prompt string, tools []contracts.Tool, options ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	// For now, call the existing method and construct a detailed response
	// TODO: Implement full detailed version that tracks token usage across all tool iterations
	content, err := c.GenerateWithTools(ctx, prompt, tools, options...)
	if err != nil {
		return nil, err
	}

	// Return a basic detailed response without usage information for now
	// This will be enhanced to track usage across all tool iterations
	return &contracts.LLMResponse{
		Content:    content,
		Model:      c.model,
		StopReason: "",
		Usage:      nil, // TODO: Implement token usage tracking for tool iterations
		Metadata: map[string]interface{}{
			"provider":   "gemini",
			"tools_used": true,
		},
	}, nil
}

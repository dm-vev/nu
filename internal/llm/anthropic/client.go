package anthropic

import (
	"context"
	"net/http"
	"time"

	"nu/internal/llm"
	"nu/internal/telemetry"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// AnthropicClient implements the LLM interface for Anthropic
type Client struct {
	APIKey              string
	Model               string
	BaseURL             string
	HTTPClient          *http.Client
	logger              telemetry.Logger
	retryExecutor       *llm.RetryExecutor
	vertexRetryExecutor *VertexRetryExecutor
	VertexConfig        *VertexConfig
	BedrockConfig       *BedrockConfig
}

// AnthropicOption represents an option for configuring the Anthropic client
type Option func(*Client)

// WithAnthropicModel sets the model for the Anthropic client
func WithModel(model string) Option {
	return func(c *Client) { c.Model = model }
}

// WithAnthropicLogger sets the logger for the Anthropic client
func WithLogger(logger telemetry.Logger) Option {
	return func(c *Client) { c.logger = logger }
}

// WithAnthropicRetry configures retry policy for the client
func WithRetry(opts ...llm.RetryOption) Option {
	return func(c *Client) {
		ctx := context.Background()
		policy := llm.NewRetryPolicy(opts...)
		c.logger.Debug(ctx, "Configuring retry", map[string]interface{}{
			"vertex_config_enabled": c.VertexConfig != nil && c.VertexConfig.Enabled,
			"vertex_config_region": func() string {
				if c.VertexConfig != nil {
					return c.VertexConfig.Region
				}
				return ""
			}(),
			"max_attempts": policy.MaximumAttempts,
		})
		if c.VertexConfig != nil && c.VertexConfig.Enabled {
			vertexPolicy := &VertexRetryPolicy{
				InitialInterval: policy.InitialInterval, BackoffCoefficient: policy.BackoffCoefficient,
				MaximumInterval: policy.MaximumInterval, MaximumAttempts: policy.MaximumAttempts,
			}
			c.vertexRetryExecutor = NewVertexRetryExecutor(c.VertexConfig, vertexPolicy)
			c.logger.Info(ctx, "Created vertex retry executor with multi-region support", map[string]interface{}{
				"region": c.VertexConfig.Region, "max_attempts": policy.MaximumAttempts,
			})
		} else {
			c.retryExecutor = llm.NewRetryExecutor(policy)
			c.logger.Info(ctx, "Created standard retry executor", map[string]interface{}{
				"max_attempts": policy.MaximumAttempts, "vertex_enabled": false,
			})
		}
	}
}

// WithAnthropicBaseURL sets the base URL for the Anthropic API
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.BaseURL = baseURL }
}

// WithAnthropicHTTPClient sets the HTTP client for the Anthropic client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.HTTPClient = httpClient }
}

// WithAnthropicVertexAI configures the client for Google Vertex AI
func WithVertexAI(region, projectID string) Option {
	return func(c *Client) {
		ctx := context.Background()
		c.logger.Debug(ctx, "Configuring Vertex AI", map[string]interface{}{
			"region": region, "projectID": projectID, "retry_executor_exists": c.retryExecutor != nil,
		})
		vertexConfig, err := NewVertexConfig(ctx, region, projectID)
		if err != nil {
			c.logger.Error(ctx, "Failed to configure Vertex AI", map[string]interface{}{
				"error": err.Error(), "region": region, "projectID": projectID,
			})
			return
		}
		c.VertexConfig = vertexConfig
		c.BaseURL = vertexConfig.GetBaseURL()
		if c.retryExecutor != nil {
			c.logger.Debug(ctx, "Creating vertex retry executor (retry executor exists)", map[string]interface{}{"region": region})
			policy := &VertexRetryPolicy{
				InitialInterval: time.Second, BackoffCoefficient: 2.0,
				MaximumInterval: time.Second * 30, MaximumAttempts: 3,
			}
			c.vertexRetryExecutor = NewVertexRetryExecutor(c.VertexConfig, policy)
			c.logger.Info(ctx, "Created vertex retry executor with multi-region support", map[string]interface{}{"region": region})
		} else {
			c.logger.Debug(ctx, "Retry executor not yet configured, vertex retry executor will be created when retry is configured", nil)
		}
		c.logger.Info(ctx, "Configured client for Vertex AI", map[string]interface{}{
			"region": region, "projectID": projectID, "baseURL": c.BaseURL,
			"vertex_retry_executor_created": c.vertexRetryExecutor != nil,
		})
	}
}

// WithAnthropicVertexAICredentials configures Vertex AI with explicit credentials
func WithVertexAICredentials(region, projectID, credentialsPath string) Option {
	return func(c *Client) {
		ctx := context.Background()
		vertexConfig, err := NewVertexConfigWithCredentials(ctx, region, projectID, credentialsPath)
		if err != nil {
			c.logger.Error(ctx, "Failed to configure Vertex AI with credentials", map[string]interface{}{
				"error": err.Error(), "region": region, "projectID": projectID, "credentialsPath": credentialsPath,
			})
			return
		}
		c.VertexConfig = vertexConfig
		c.BaseURL = vertexConfig.GetBaseURL()
		c.logger.Info(ctx, "Configured client for Vertex AI with credentials", map[string]interface{}{
			"region": region, "projectID": projectID, "credentialsPath": credentialsPath, "baseURL": c.BaseURL,
		})
	}
}

// WithAnthropicGoogleApplicationCredentials configures Vertex AI with explicit credentials content
func WithGoogleApplicationCredentials(region, projectID, credentialsContent string) Option {
	return func(c *Client) {
		ctx := context.Background()
		vertexConfig, err := NewVertexConfigWithCredentialsContent(ctx, region, projectID, credentialsContent)
		if err != nil {
			c.logger.Error(ctx, "Failed to configure Vertex AI with credentials content", map[string]interface{}{
				"error": err.Error(), "region": region, "projectID": projectID,
			})
			return
		}
		c.VertexConfig = vertexConfig
		c.BaseURL = vertexConfig.GetBaseURL()
		c.logger.Info(ctx, "Configured client for Vertex AI with credentials content", map[string]interface{}{
			"region": region, "projectID": projectID, "baseURL": c.BaseURL,
		})
	}
}

// WithAnthropicBedrockAWSConfig configures Bedrock with an existing AWS config
func WithBedrockAWSConfig(awsConfig aws.Config) Option {
	return func(c *Client) {
		ctx := context.Background()
		bedrockConfig, err := NewBedrockConfigWithAWSConfig(ctx, awsConfig)
		if err != nil {
			c.logger.Error(ctx, "Failed to configure Bedrock with AWS config", map[string]interface{}{
				"error": err.Error(), "region": awsConfig.Region,
			})
			return
		}
		c.BedrockConfig = bedrockConfig
		c.logger.Info(ctx, "Configured client for AWS Bedrock with AWS config", map[string]interface{}{"region": awsConfig.Region})
	}
}

// NewAnthropic creates a new Anthropic client
func NewClient(apiKey string, options ...Option) *Client {
	client := &Client{
		APIKey: apiKey, Model: Claude37Sonnet, BaseURL: "https://api.anthropic.com",
		HTTPClient: &http.Client{Timeout: 30 * time.Minute}, logger: telemetry.NewLogger(),
	}
	for _, option := range options {
		option(client)
	}
	if client.VertexConfig != nil && client.VertexConfig.Enabled && client.retryExecutor != nil && client.vertexRetryExecutor == nil {
		client.logger.Error(context.TODO(), "Vertex AI configured with retry but vertex executor not created. This indicates option ordering issue - WithRetry should come after WithVertexAI.", map[string]interface{}{
			"vertex_config_enabled": true, "retry_executor_exists": true, "vertex_retry_executor_exists": false,
		})
	}
	if client.Model == "" {
		client.logger.Warn(context.TODO(), "No model specified, model must be explicitly set with WithModel", nil)
	}
	return client
}

// Name implements contracts.LLM.Name
func (c *Client) Name() string { return "anthropic" }

// SupportsStreaming implements contracts.LLM.SupportsStreaming
func (c *Client) SupportsStreaming() bool { return true }

// GetModel returns the model name being used
func (c *Client) GetModel() string { return c.Model }

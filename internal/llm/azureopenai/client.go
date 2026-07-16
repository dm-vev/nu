package azureopenai

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/llm"
	"nu/internal/telemetry"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// Define a custom type for context keys to avoid collisions
type azureOpenAIContextKey string

// Define constants for context keys
const azureOpenAIOrganizationKey azureOpenAIContextKey = "organization"

// AzureOpenAIClient implements the LLM interface for Azure OpenAI
type Client struct {
	Client          openai.Client
	ChatService     openai.ChatService
	ResponseService openai.Client
	Model           string
	apiKey          string
	baseURL         string
	apiVersion      string
	deployment      string
	region          string
	resourceName    string
	logger          telemetry.Logger
	retryExecutor   *llm.RetryExecutor
}

// recreateClients recreates the OpenAI clients with current configuration
func (c *Client) recreateClients() {
	// Build the Azure OpenAI endpoint URL
	// If baseURL is provided, use it directly, otherwise construct from region and resource name
	var azureURL string
	if c.baseURL != "" {
		// Use provided baseURL (e.g., https://your-resource.openai.azure.com)
		azureURL = fmt.Sprintf("%s/openai/deployments/%s", strings.TrimSuffix(c.baseURL, "/"), c.deployment)
	} else if c.region != "" && c.resourceName != "" {
		// Construct URL from region and resource name
		azureURL = fmt.Sprintf("https://%s.openai.azure.com/openai/deployments/%s", c.resourceName, c.deployment)
		c.baseURL = fmt.Sprintf("https://%s.openai.azure.com", c.resourceName)
	} else {
		c.logger.Error(context.Background(), "Either baseURL or both region and resourceName must be provided", nil)
		return
	}

	options := []option.RequestOption{
		option.WithAPIKey(c.apiKey),
		option.WithBaseURL(azureURL),
	}

	// Add API version as query parameter for Azure OpenAI
	if c.apiVersion != "" {
		options = append(options, option.WithQuery("api-version", c.apiVersion))
	}

	c.Client = openai.NewClient(options...)
	c.ChatService = openai.NewChatService(options...)
	c.ResponseService = openai.NewClient(options...)
}

// NewAzureOpenAI creates a new Azure OpenAI client
// You can provide either:
// 1. baseURL (e.g., "https://your-resource.openai.azure.com") - traditional approach
// 2. region and resourceName via options - newer approach
func NewClient(apiKey, baseURL, deployment string, options ...Option) *Client {
	// Create client with default options
	client := &Client{
		Model:      deployment, // In Azure OpenAI, deployment name is the model identifier
		apiKey:     apiKey,
		baseURL:    baseURL,
		deployment: deployment,
		apiVersion: "2024-08-01-preview", // Default API version (required for structured output)
		logger:     telemetry.NewLogger(),
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	// Create the OpenAI clients with Azure configuration
	client.recreateClients()

	return client
}

// NewClientFromRegion creates a new Azure OpenAI client using region and resource name
// This is the recommended approach for new implementations
func NewClientFromRegion(apiKey, region, resourceName, deployment string, options ...Option) *Client {
	// Create client with region-based configuration
	client := &Client{
		Model:        deployment, // In Azure OpenAI, deployment name is the model identifier
		apiKey:       apiKey,
		region:       region,
		resourceName: resourceName,
		deployment:   deployment,
		apiVersion:   "2024-08-01-preview", // Default API version (required for structured output)
		logger:       telemetry.NewLogger(),
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	// Create the OpenAI clients with Azure configuration
	client.recreateClients()

	return client
}

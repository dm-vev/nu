package gemini

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials"
	"google.golang.org/genai"

	"github.com/dm-vev/nu/telemetry"
)

// NewClient creates a Gemini client.
func NewClient(ctx context.Context, options ...Option) (*Client, error) {
	// Create client with default options
	defaultThinking := DefaultThinkingConfig()
	client := &Client{
		model:          DefaultModel,
		backend:        genai.BackendGeminiAPI,
		location:       "us-central1", // Default Vertex AI location
		logger:         telemetry.NewLogger(),
		thinkingConfig: &defaultThinking,
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	// Validate that only one credential type is provided
	credentialTypesProvided := 0
	if client.credentialsFile != "" {
		credentialTypesProvided++
	}
	if len(client.credentialsJSON) > 0 {
		credentialTypesProvided++
	}

	if credentialTypesProvided > 1 {
		return nil, fmt.Errorf("only one credential type can be provided: choose between WithCredentialsFile or WithCredentialsJSON")
	}

	// If an existing client was injected, use it
	if client.genaiClient != nil {
		return client, nil
	}

	// Create the genai client if not already provided
	if client.genaiClient == nil {
		config := &genai.ClientConfig{
			Backend: client.backend,
		}

		// Configure based on backend type
		switch client.backend {
		case genai.BackendGeminiAPI:
			if client.apiKey == "" {
				return nil, fmt.Errorf("API key is required for Gemini API backend")
			}
			config.APIKey = client.apiKey
		case genai.BackendVertexAI:
			// Validate that at least one authentication method is provided
			if client.projectID == "" && client.credentialsFile == "" && len(client.credentialsJSON) == 0 && client.apiKey == "" {
				return nil, fmt.Errorf("project ID, credentials file, credentials JSON, or API key are required for Vertex AI backend")
			}

			// Handle service account credentials
			if client.credentialsFile != "" {
				// Handle service account credentials from file
				creds, err := credentials.DetectDefault(&credentials.DetectOptions{
					CredentialsFile: client.credentialsFile,
					Scopes: []string{
						"https://www.googleapis.com/auth/cloud-platform",
					},
				})
				if err != nil {
					return nil, fmt.Errorf("failed to load credentials from file: %w", err)
				}
				config.Credentials = creds
			} else if len(client.credentialsJSON) > 0 {
				// Handle service account credentials from JSON
				creds, err := credentials.DetectDefault(&credentials.DetectOptions{
					CredentialsJSON: client.credentialsJSON,
					Scopes: []string{
						"https://www.googleapis.com/auth/cloud-platform",
					},
				})
				if err != nil {
					return nil, fmt.Errorf("failed to load credentials from JSON: %w", err)
				}
				config.Credentials = creds
			}

			// Set project and location if provided
			if client.projectID != "" {
				config.Project = client.projectID
				config.Location = client.location
			}

			// Set API key if provided (alternative authentication method)
			if client.apiKey != "" {
				config.APIKey = client.apiKey
			}
		}

		genaiClient, err := genai.NewClient(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}

		client.genaiClient = genaiClient
	}

	return client, nil
}

package providers

import (
	"context"
	"fmt"

	"google.golang.org/genai"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm/gemini"
)

// createGeminiClient creates a Google Gemini LLM client
// Supports both API key and Vertex AI authentication
func createGeminiClient(cfg *config.LLMProviderYAML) (contracts.LLM, error) {
	var options []gemini.Option

	// Set model - use config model or fallback to GEMINI_MODEL env var
	model := config.ExpandEnv(cfg.Model)
	if model == "" {
		model = getConfigString(cfg.Config, "model")
	}
	if model == "" {
		model = config.GetEnvValue("GEMINI_MODEL")
	}
	if model != "" {
		options = append(options, gemini.WithModel(model))
	}

	// Check for Vertex AI credentials first (preferred for production)
	googleCreds := getConfigString(cfg.Config, "google_application_credentials")
	if googleCreds == "" {
		googleCreds = config.GetEnvValue("VERTEX_AI_GOOGLE_APPLICATION_CREDENTIALS_CONTENT")
	}

	projectID := getConfigString(cfg.Config, "project_id")
	if projectID == "" {
		projectID = getConfigString(cfg.Config, "project")
	}
	if projectID == "" {
		projectID = config.GetEnvValue("VERTEX_AI_PROJECT")
	}

	// Use Vertex AI if credentials and project are available
	if googleCreds != "" && projectID != "" {
		location := getConfigString(cfg.Config, "location")
		if location == "" {
			location = config.GetEnvValue("VERTEX_AI_REGION")
		}
		if location == "" {
			location = "us-central1"
		}

		// Parse credentials (supports base64 encoded, file path, or raw JSON)
		credentialsJSON, err := ParseGoogleCredentials(googleCreds)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Google credentials: %w", err)
		}

		options = append(options,
			gemini.WithBackend(genai.BackendVertexAI),
			gemini.WithCredentialsJSON([]byte(credentialsJSON)),
			gemini.WithProjectID(projectID),
			gemini.WithLocation(location),
		)
	} else {
		// Fall back to API key authentication
		apiKey := getConfigString(cfg.Config, "api_key")
		if apiKey == "" {
			apiKey = config.GetEnvValue("GEMINI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("credentials required for Gemini provider: set GEMINI_API_KEY or Vertex AI credentials (VERTEX_AI_PROJECT + VERTEX_AI_GOOGLE_APPLICATION_CREDENTIALS_CONTENT)")
		}
		options = append(options, gemini.WithAPIKey(apiKey))
	}

	// Create context for client initialization
	ctx := context.Background()
	client, err := gemini.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return client, nil
}

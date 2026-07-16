package providers

import (
	"fmt"
	"strings"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm/anthropic"
)

// createAnthropicClient creates an Anthropic LLM client
func createAnthropicClient(cfg *config.LLMProviderYAML) (contracts.LLM, error) {
	var options []anthropic.Option
	var apiKey string

	// Check for Vertex AI configuration first (preferred method)
	vertexProject := getConfigString(cfg.Config, "vertex_ai_project")
	if vertexProject == "" {
		vertexProject = config.GetEnvValue("VERTEX_AI_PROJECT")
	}

	// Use Vertex AI if configured
	if vertexProject != "" {
		// Validate project ID format (basic validation)
		if strings.TrimSpace(vertexProject) == "" {
			return nil, fmt.Errorf("vertex_ai_project cannot be empty or whitespace-only")
		}
		// Check for both vertex_ai_region and vertex_ai_location for backward compatibility
		location := getConfigString(cfg.Config, "vertex_ai_region")
		if location == "" {
			location = getConfigString(cfg.Config, "vertex_ai_location")
		}
		if location == "" {
			location = config.GetEnvValue("VERTEX_AI_REGION")
		}
		if location == "" {
			location = config.GetEnvValue("VERTEX_AI_LOCATION")
		}
		if location == "" {
			location = "us-central1" // Default location
		}

		// Check if explicit credentials are provided
		if creds := getConfigString(cfg.Config, "google_application_credentials"); creds != "" {
			// Parse credentials - could be file path, base64, or raw JSON
			credContent, err := ParseGoogleCredentials(creds)
			if err != nil {
				return nil, fmt.Errorf("failed to parse google_application_credentials for Vertex AI project %s: %w", vertexProject, err)
			}
			options = append(options, anthropic.WithGoogleApplicationCredentials(location, vertexProject, credContent))
		} else {
			// Use default ADC
			options = append(options, anthropic.WithVertexAI(location, vertexProject))
		}

		// Use placeholder API key for Vertex AI
		apiKey = "vertex-ai"
	} else {
		// Fallback to Anthropic API with API key
		apiKey = getConfigString(cfg.Config, "api_key")
		if apiKey == "" {
			apiKey = config.GetEnvValue("ANTHROPIC_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("api_key is required for Anthropic provider (set ANTHROPIC_API_KEY or config.api_key) or configure Vertex AI (set VERTEX_AI_PROJECT and optionally VERTEX_AI_REGION)")
		}
	}

	// Set model - use config model or fallback to ANTHROPIC_MODEL env var
	model := config.ExpandEnv(cfg.Model)
	if model == "" {
		model = getConfigString(cfg.Config, "model")
	}
	if model == "" {
		model = config.GetEnvValue("ANTHROPIC_MODEL")
	}
	if model != "" {
		options = append(options, anthropic.WithModel(model))
	}

	// Set base URL if provided (for custom endpoints)
	if baseURL := getConfigString(cfg.Config, "base_url"); baseURL != "" {
		options = append(options, anthropic.WithBaseURL(baseURL))
	}

	return anthropic.NewClient(apiKey, options...), nil
}

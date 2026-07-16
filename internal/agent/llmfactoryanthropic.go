package agent

import (
	"fmt"
	"strings"

	agentconfig "nu/internal/agent/config"
	"nu/internal/contracts"
	"nu/internal/llm/anthropic"
)

// createAnthropicClient creates an Anthropic LLM client
func createAnthropicClient(config *agentconfig.LLMProviderYAML) (contracts.LLM, error) {
	var options []anthropic.Option
	var apiKey string

	// Check for Vertex AI configuration first (preferred method)
	vertexProject := getConfigString(config.Config, "vertex_ai_project")
	if vertexProject == "" {
		vertexProject = agentconfig.GetEnvValue("VERTEX_AI_PROJECT")
	}

	// Use Vertex AI if configured
	if vertexProject != "" {
		// Validate project ID format (basic validation)
		if strings.TrimSpace(vertexProject) == "" {
			return nil, fmt.Errorf("vertex_ai_project cannot be empty or whitespace-only")
		}
		// Check for both vertex_ai_region and vertex_ai_location for backward compatibility
		location := getConfigString(config.Config, "vertex_ai_region")
		if location == "" {
			location = getConfigString(config.Config, "vertex_ai_location")
		}
		if location == "" {
			location = agentconfig.GetEnvValue("VERTEX_AI_REGION")
		}
		if location == "" {
			location = agentconfig.GetEnvValue("VERTEX_AI_LOCATION")
		}
		if location == "" {
			location = "us-central1" // Default location
		}

		// Check if explicit credentials are provided
		if creds := getConfigString(config.Config, "google_application_credentials"); creds != "" {
			// Parse credentials - could be file path, base64, or raw JSON
			credContent, err := parseGoogleCredentials(creds)
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
		apiKey = getConfigString(config.Config, "api_key")
		if apiKey == "" {
			apiKey = agentconfig.GetEnvValue("ANTHROPIC_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("api_key is required for Anthropic provider (set ANTHROPIC_API_KEY or config.api_key) or configure Vertex AI (set VERTEX_AI_PROJECT and optionally VERTEX_AI_REGION)")
		}
	}

	// Set model - use config model or fallback to ANTHROPIC_MODEL env var
	model := agentconfig.ExpandEnv(config.Model)
	if model == "" {
		model = getConfigString(config.Config, "model")
	}
	if model == "" {
		model = agentconfig.GetEnvValue("ANTHROPIC_MODEL")
	}
	if model != "" {
		options = append(options, anthropic.WithModel(model))
	}

	// Set base URL if provided (for custom endpoints)
	if baseURL := getConfigString(config.Config, "base_url"); baseURL != "" {
		options = append(options, anthropic.WithBaseURL(baseURL))
	}

	return anthropic.NewClient(apiKey, options...), nil
}

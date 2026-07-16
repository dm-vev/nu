package image

import (
	"fmt"
	"os"

	"google.golang.org/genai"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/agent/providers"
	"github.com/dm-vev/nu/internal/llm/gemini"
)

type credentials struct {
	googleCredentials string
	projectID         string
	location          string
	apiKey            string
}

func resolveCredentials(config *config.ImageGenerationYAML) credentials {
	result := credentials{}
	if config.Config != nil {
		result.googleCredentials, _ = config.Config["google_application_credentials"].(string)
		result.projectID, _ = config.Config["project_id"].(string)
		result.location, _ = config.Config["location"].(string)
		result.apiKey, _ = config.Config["api_key"].(string)
	}
	if result.googleCredentials == "" {
		result.googleCredentials = os.Getenv("VERTEX_AI_GOOGLE_APPLICATION_CREDENTIALS_CONTENT")
	}
	if result.projectID == "" {
		result.projectID = os.Getenv("VERTEX_AI_PROJECT")
	}
	if result.location == "" {
		result.location = os.Getenv("VERTEX_AI_REGION")
	}
	if result.location == "" {
		result.location = "us-central1"
	}
	if result.apiKey == "" {
		result.apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if result.apiKey == "" {
		result.apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	return result
}

func (c credentials) geminiOptions() ([]gemini.Option, error) {
	if c.googleCredentials != "" && c.projectID != "" {
		credentialsJSON, err := providers.ParseGoogleCredentials(c.googleCredentials)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Google credentials for image generation: %w", err)
		}
		return []gemini.Option{
			gemini.WithBackend(genai.BackendVertexAI),
			gemini.WithCredentialsJSON([]byte(credentialsJSON)),
			gemini.WithProjectID(c.projectID),
			gemini.WithLocation(c.location),
		}, nil
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("credentials required for image generation: set GEMINI_API_KEY or Vertex AI credentials (VERTEX_AI_PROJECT + VERTEX_AI_GOOGLE_APPLICATION_CREDENTIALS_CONTENT)")
	}
	return []gemini.Option{gemini.WithAPIKey(c.apiKey)}, nil
}

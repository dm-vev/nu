package embedding

import (
	"google.golang.org/genai"

	"nu/internal/telemetry"
)

// GeminiEmbedderOption represents an option for configuring the Gemini embedder
type GeminiEmbedderOption func(*GeminiEmbedder)

// WithGeminiModel sets the embedding model for the Gemini embedder
func WithGeminiModel(model string) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.model = model
		e.config.Model = model
	}
}

// WithGeminiAPIKey sets the API key for Gemini API backend
func WithGeminiAPIKey(apiKey string) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.apiKey = apiKey
	}
}

// WithGeminiBackend sets the backend for the Gemini embedder
func WithGeminiBackend(backend genai.Backend) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.backend = backend
	}
}

// WithGeminiProjectID sets the GCP project ID for Vertex AI backend
func WithGeminiProjectID(projectID string) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.projectID = projectID
	}
}

// WithGeminiLocation sets the GCP location for Vertex AI backend
func WithGeminiLocation(location string) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.location = location
	}
}

// WithGeminiCredentialsFile sets the path to a service account key file for Vertex AI authentication
func WithGeminiCredentialsFile(credentialsFile string) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.credentialsFile = credentialsFile
	}
}

// WithGeminiCredentialsJSON sets the service account key JSON bytes for Vertex AI authentication
func WithGeminiCredentialsJSON(credentialsJSON []byte) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.credentialsJSON = credentialsJSON
	}
}

// WithGeminiLogger sets the logger for the Gemini embedder
func WithGeminiLogger(logger telemetry.Logger) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.logger = logger
	}
}

// WithGeminiConfig sets the embedding configuration for the Gemini embedder
func WithGeminiConfig(config Config) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.config = config
		if config.Model != "" {
			e.model = config.Model
		}
	}
}

// WithGeminiTaskType sets the task type for better embedding optimization
// Valid values: "RETRIEVAL_QUERY", "RETRIEVAL_DOCUMENT", "SEMANTIC_SIMILARITY",
// "CLASSIFICATION", "CLUSTERING", "QUESTION_ANSWERING", "FACT_VERIFICATION"
func WithGeminiTaskType(taskType string) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.taskType = taskType
	}
}

// WithGeminiClient injects an already initialized genai.Client
func WithGeminiClient(existing *genai.Client) GeminiEmbedderOption {
	return func(e *GeminiEmbedder) {
		e.client = existing
	}
}

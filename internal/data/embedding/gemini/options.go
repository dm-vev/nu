package gemini

import (
	"google.golang.org/genai"

	"github.com/dm-vev/nu/internal/data/embedding"
	"github.com/dm-vev/nu/telemetry"
)

// Option represents an option for configuring the Gemini embedder
type Option func(*Client)

// WithModel sets the embedding model for the Gemini embedder
func WithModel(model string) Option {
	return func(e *Client) {
		e.model = model
		e.config.Model = model
	}
}

// WithAPIKey sets the API key for Gemini API backend
func WithAPIKey(apiKey string) Option {
	return func(e *Client) {
		e.apiKey = apiKey
	}
}

// WithBackend sets the backend for the Gemini embedder
func WithBackend(backend genai.Backend) Option {
	return func(e *Client) {
		e.backend = backend
	}
}

// WithProjectID sets the GCP project ID for Vertex AI backend
func WithProjectID(projectID string) Option {
	return func(e *Client) {
		e.projectID = projectID
	}
}

// WithLocation sets the GCP location for Vertex AI backend
func WithLocation(location string) Option {
	return func(e *Client) {
		e.location = location
	}
}

// WithCredentialsFile sets the path to a service account key file for Vertex AI authentication
func WithCredentialsFile(credentialsFile string) Option {
	return func(e *Client) {
		e.credentialsFile = credentialsFile
	}
}

// WithCredentialsJSON sets the service account key JSON bytes for Vertex AI authentication
func WithCredentialsJSON(credentialsJSON []byte) Option {
	return func(e *Client) {
		e.credentialsJSON = credentialsJSON
	}
}

// WithLogger sets the logger for the Gemini embedder
func WithLogger(logger telemetry.Logger) Option {
	return func(e *Client) {
		e.logger = logger
	}
}

// WithConfig sets the embedding configuration for the Gemini embedder
func WithConfig(config embedding.Config) Option {
	return func(e *Client) {
		e.config = config
		if config.Model != "" {
			e.model = config.Model
		}
	}
}

// WithTaskType sets the task type for better embedding optimization
// Valid values: "RETRIEVAL_QUERY", "RETRIEVAL_DOCUMENT", "SEMANTIC_SIMILARITY",
// "CLASSIFICATION", "CLUSTERING", "QUESTION_ANSWERING", "FACT_VERIFICATION"
func WithTaskType(taskType string) Option {
	return func(e *Client) {
		e.taskType = taskType
	}
}

// WithClient injects an already initialized genai.Client
func WithClient(existing *genai.Client) Option {
	return func(e *Client) {
		e.client = existing
	}
}

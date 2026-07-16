package embedding

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials"
	"google.golang.org/genai"

	"nu/internal/telemetry"
)

// Gemini embedding model constants
const (
	// ModelTextEmbedding004 is the latest text embedding model (768 dimensions)
	ModelTextEmbedding004 = "text-embedding-004"

	// ModelTextEmbedding005 is the newest text embedding model (768 dimensions)
	ModelTextEmbedding005 = "text-embedding-005"

	// ModelTextMultilingualEmbedding002 is for multilingual text (768 dimensions)
	ModelTextMultilingualEmbedding002 = "text-multilingual-embedding-002"

	// DefaultGeminiEmbeddingModel is the default embedding model
	DefaultGeminiEmbeddingModel = ModelTextEmbedding004
)

// GeminiEmbedder implements embedding generation using Google Gemini/Vertex AI API
type GeminiEmbedder struct {
	client          *genai.Client
	model           string
	config          Config
	backend         genai.Backend
	projectID       string
	location        string
	credentialsFile string
	credentialsJSON []byte
	apiKey          string
	logger          telemetry.Logger
	taskType        string // Optional task type for better embeddings
}

// NewGemini creates a Gemini embedder with the provided options.
func NewGemini(ctx context.Context, options ...GeminiEmbedderOption) (*GeminiEmbedder, error) {
	// Create embedder with default options
	embedder := &GeminiEmbedder{
		model:    DefaultGeminiEmbeddingModel,
		backend:  genai.BackendGeminiAPI,
		location: "us-central1", // Default Vertex AI location
		logger:   telemetry.NewLogger(),
		config:   DefaultGeminiEmbeddingConfig(""),
	}

	// Apply options
	for _, option := range options {
		option(embedder)
	}

	// Update config model if not set
	if embedder.config.Model == "" {
		embedder.config.Model = embedder.model
	}

	// Validate that only one credential type is provided
	credentialTypesProvided := 0
	if embedder.credentialsFile != "" {
		credentialTypesProvided++
	}
	if len(embedder.credentialsJSON) > 0 {
		credentialTypesProvided++
	}

	if credentialTypesProvided > 1 {
		return nil, fmt.Errorf("only one credential type can be provided: choose between WithGeminiCredentialsFile or WithGeminiCredentialsJSON")
	}

	// If an existing client was injected, use it
	if embedder.client != nil {
		return embedder, nil
	}

	// Create the genai client
	clientConfig := &genai.ClientConfig{
		Backend: embedder.backend,
	}

	// Configure based on backend type
	switch embedder.backend {
	case genai.BackendGeminiAPI:
		if embedder.apiKey == "" {
			return nil, fmt.Errorf("API key is required for Gemini API backend")
		}
		clientConfig.APIKey = embedder.apiKey

	case genai.BackendVertexAI:
		// Validate that at least one authentication method is provided
		if embedder.projectID == "" && embedder.credentialsFile == "" && len(embedder.credentialsJSON) == 0 && embedder.apiKey == "" {
			return nil, fmt.Errorf("project ID, credentials file, credentials JSON, or API key are required for Vertex AI backend")
		}

		// Handle service account credentials
		if embedder.credentialsFile != "" {
			creds, err := credentials.DetectDefault(&credentials.DetectOptions{
				CredentialsFile: embedder.credentialsFile,
				Scopes: []string{
					"https://www.googleapis.com/auth/cloud-platform",
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to load credentials from file: %w", err)
			}
			clientConfig.Credentials = creds
		} else if len(embedder.credentialsJSON) > 0 {
			creds, err := credentials.DetectDefault(&credentials.DetectOptions{
				CredentialsJSON: embedder.credentialsJSON,
				Scopes: []string{
					"https://www.googleapis.com/auth/cloud-platform",
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to load credentials from JSON: %w", err)
			}
			clientConfig.Credentials = creds
		}

		// Set project and location if provided
		if embedder.projectID != "" {
			clientConfig.Project = embedder.projectID
			clientConfig.Location = embedder.location
		}

		// Set API key if provided (alternative authentication method)
		if embedder.apiKey != "" {
			clientConfig.APIKey = embedder.apiKey
		}
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	embedder.client = client

	return embedder, nil
}

// DefaultGeminiEmbeddingConfig returns a default configuration for Gemini embedding generation
func DefaultGeminiEmbeddingConfig(model string) Config {
	if model == "" {
		model = DefaultGeminiEmbeddingModel
	}

	return Config{
		Model:               model,
		Dimensions:          768, // Default dimensions for Gemini embedding models
		EncodingFormat:      "float",
		SimilarityMetric:    "cosine",
		SimilarityThreshold: 0.0,
	}
}

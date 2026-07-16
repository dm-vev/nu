// Package vector provides a Weaviate-backed vector store.
package vector

import (
	"context"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/auth"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/embedding"
	"github.com/dm-vev/nu/telemetry"
)

// Config configures vector storage backends.
type Config = contracts.VectorStoreConfig

// Store implements contracts.VectorStore using Weaviate.
type Store struct {
	client         *weaviate.Client
	classPrefix    string
	embedder       embedding.Client
	distanceMetric string
	logger         telemetry.Logger
}

// Option configures a Store.
type Option func(*Store)

// WithClassPrefix sets the class prefix for the Weaviate store
func WithClassPrefix(prefix string) Option {
	return func(s *Store) {
		s.classPrefix = prefix
	}
}

// WithEmbedder sets the embedder for the Weaviate store
func WithEmbedder(embedder embedding.Client) Option {
	return func(s *Store) {
		s.embedder = embedder
	}
}

// WithDistanceMetric sets the distance metric for the Weaviate store
func WithDistanceMetric(metric string) Option {
	return func(s *Store) {
		s.distanceMetric = metric
	}
}

// WithLogger sets the logger for the Weaviate store
func WithLogger(logger telemetry.Logger) Option {
	return func(s *Store) {
		s.logger = logger
	}
}

// NewStore creates a Weaviate vector store.
func NewStore(config *Config, options ...Option) *Store {
	// Create store with default options
	store := &Store{
		classPrefix:    "Document",
		distanceMetric: "cosine",
		logger:         telemetry.NewLogger(),
	}

	// Apply options
	for _, option := range options {
		option(store)
	}

	// Create Weaviate client
	cfg := weaviate.Config{
		Host:   config.Host,
		Scheme: config.Scheme,
	}

	// Add API key if provided - use AuthConfig for proper Weaviate Cloud support
	if config.APIKey != "" {
		cfg.AuthConfig = auth.ApiKey{Value: config.APIKey}
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		store.logger.Error(context.Background(), "Failed to create Weaviate client", map[string]interface{}{"error": err.Error()})
		return nil
	}

	store.client = client

	return store
}

// getClassName returns the class name
// Uses metadata-based multi-tenancy (single class, orgId as field) instead of class proliferation
func (s *Store) getClassName(ctx context.Context, class string) (string, error) {
	// If class is provided, use it; otherwise use default
	if class == "" {
		class = s.classPrefix
	}

	// Always return the base class name
	// Multi-tenancy is handled via orgId field filtering, not separate classes
	s.logger.Debug(ctx, "Using single class with metadata-based multi-tenancy", map[string]interface{}{
		"class": class,
	})
	return class, nil
}

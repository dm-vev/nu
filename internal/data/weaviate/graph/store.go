// Package graph provides a Weaviate-based implementation of the GraphRAG interface.
//
// This implementation stores entities and relationships in separate Weaviate collections,
// using vector embeddings for semantic search and metadata filtering for graph traversal.
package graph

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/auth"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/embedding"
	"github.com/dm-vev/nu/telemetry"
)

// Store implements contracts.GraphRAGStore using Weaviate.
type Store struct {
	client      *weaviate.Client
	classPrefix string
	embedder    embedding.Client
	logger      telemetry.Logger
	tenant      string
	schema      *contracts.GraphSchema
}

// StoreConfig configures a Weaviate GraphRAG store.
type StoreConfig struct {
	// Host is the hostname of the Weaviate server (e.g., "localhost:8080")
	Host string

	// Scheme is the URL scheme ("http" or "https")
	Scheme string

	// APIKey is the authentication key for Weaviate Cloud
	APIKey string

	// ClassPrefix is the prefix for entity/relationship collections (default: "Graph")
	ClassPrefix string
}

// StoreOption configures a Store.
type StoreOption func(*Store)

// WithClassPrefix sets the class prefix for entity/relationship collections.
func WithClassPrefix(prefix string) StoreOption {
	return func(s *Store) {
		s.classPrefix = prefix
	}
}

// WithEmbedder sets the embedder for generating vectors.
func WithEmbedder(embedder embedding.Client) StoreOption {
	return func(s *Store) {
		s.embedder = embedder
	}
}

// WithLogger sets the logger for the store.
func WithLogger(logger telemetry.Logger) StoreOption {
	return func(s *Store) {
		s.logger = logger
	}
}

// WithDefaultTenant sets the default tenant.
func WithDefaultTenant(tenant string) StoreOption {
	return func(s *Store) {
		s.tenant = tenant
	}
}

// WithSchema sets an initial schema for the store.
func WithSchema(schema *contracts.GraphSchema) StoreOption {
	return func(s *Store) {
		s.schema = schema
	}
}

// NewStore creates a Weaviate GraphRAG store.
func NewStore(config *StoreConfig, options ...StoreOption) (*Store, error) {
	if config == nil {
		config = &StoreConfig{
			Host:        "localhost:8080",
			Scheme:      "http",
			ClassPrefix: "Graph",
		}
	}

	store := &Store{
		classPrefix: "Graph",
		logger:      telemetry.NewLogger(),
	}

	// Override classPrefix from config if provided
	if config.ClassPrefix != "" {
		store.classPrefix = config.ClassPrefix
	}

	// Apply options
	for _, option := range options {
		option(store)
	}

	// Create Weaviate client configuration
	cfg := weaviate.Config{
		Host:   config.Host,
		Scheme: config.Scheme,
	}

	// Add API key authentication if provided
	if config.APIKey != "" {
		cfg.AuthConfig = auth.ApiKey{Value: config.APIKey}
	}

	// Create Weaviate client
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	store.client = client

	// Ensure schema exists
	if err := store.ensureSchema(context.Background()); err != nil {
		store.logger.Warn(context.Background(), "Failed to ensure schema, collections may need to be created manually", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return store, nil
}

// getEntityClassName returns the entity collection name.
func (s *Store) getEntityClassName() string {
	return s.classPrefix + "Entity"
}

// getRelationshipClassName returns the relationship collection name.
func (s *Store) getRelationshipClassName() string {
	return s.classPrefix + "Relationship"
}

// SetTenant sets the current tenant for multi-tenancy operations.
func (s *Store) SetTenant(tenant string) {
	s.tenant = tenant
}

// GetTenant returns the current tenant.
func (s *Store) GetTenant() string {
	return s.tenant
}

// Close closes the store connection.
// Note: Weaviate client doesn't require explicit closing.
func (s *Store) Close() error {
	return nil
}

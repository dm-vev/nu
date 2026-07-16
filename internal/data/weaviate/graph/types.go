// Package graph provides GraphRAG types backed by Weaviate.
//
// GraphRAG extends traditional RAG by leveraging knowledge graphs for enhanced context retrieval.
// Unlike pure vector search, GraphRAG maintains explicit relationships between entities, enabling:
//
//   - Relationship-aware retrieval
//   - Multi-hop graph traversal
//   - Community-based global search
//   - Entity extraction and knowledge building
package graph

import (
	"github.com/dm-vev/nu/contracts"
)

// Type aliases for convenience; canonical graph types live in contracts.
type (
	// Entity represents a node in the knowledge graph
	Entity = contracts.Entity

	// Relationship represents an edge connecting two entities
	Relationship = contracts.Relationship

	// SearchResult represents a search result from the knowledge graph.
	SearchResult = contracts.GraphSearchResult

	// Context represents context around a central entity from graph traversal.
	Context = contracts.GraphContext

	// Path represents a path between two entities.
	Path = contracts.GraphPath

	// ExtractionResult contains extracted entities and relationships from text
	ExtractionResult = contracts.ExtractionResult

	// Schema defines the structure of the knowledge graph.
	Schema = contracts.GraphSchema

	// EntityTypeSchema defines an entity type in the schema
	EntityTypeSchema = contracts.EntityTypeSchema

	// RelationshipTypeSchema defines a relationship type in the schema
	RelationshipTypeSchema = contracts.RelationshipTypeSchema

	// PropertySchema defines a property in the schema
	PropertySchema = contracts.PropertySchema

	// StoreOptions contains options for storing graph data.
	StoreOptions = contracts.GraphStoreOptions

	// SearchOptions contains options for searching graph data.
	SearchOptions = contracts.GraphSearchOptions

	// ExtractionOptions contains options for extraction operations
	ExtractionOptions = contracts.ExtractionOptions

	// RelationshipDirection specifies the direction for relationship queries
	RelationshipDirection = contracts.RelationshipDirection

	// SearchMode specifies the type of search to perform.
	SearchMode = contracts.GraphSearchMode
)

// Direction constants
const (
	DirectionOutgoing = contracts.DirectionOutgoing
	DirectionIncoming = contracts.DirectionIncoming
	DirectionBoth     = contracts.DirectionBoth
)

// Search mode constants
const (
	SearchModeVector  = contracts.SearchModeVector
	SearchModeKeyword = contracts.SearchModeKeyword
	SearchModeHybrid  = contracts.SearchModeHybrid
)

// Config holds configuration for GraphRAG providers.
type Config struct {
	// Provider is the backend provider ("weaviate", "neo4j")
	Provider string `json:"provider" yaml:"provider"`

	// Host is the hostname of the backend server
	Host string `json:"host" yaml:"host"`

	// Scheme is the URL scheme (http or https)
	Scheme string `json:"scheme" yaml:"scheme"`

	// APIKey is the authentication key
	APIKey string `json:"api_key" yaml:"api_key"`

	// ClassPrefix is the prefix for collection/class names
	ClassPrefix string `json:"class_prefix" yaml:"class_prefix"`

	// Schema is the optional schema definition
	Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// DefaultConfig returns a default GraphRAG configuration.
func DefaultConfig() *Config {
	return &Config{
		Provider:    "weaviate",
		Host:        "localhost:8080",
		Scheme:      "http",
		ClassPrefix: "Graph",
	}
}

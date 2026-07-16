package contracts

import (
	"context"
)

// GraphRAGStore defines the interface for graph-based retrieval-augmented generation
type GraphRAGStore interface {
	// Entity CRUD operations
	StoreEntities(ctx context.Context, entities []Entity, options ...GraphStoreOption) error
	GetEntity(ctx context.Context, id string, options ...GraphStoreOption) (*Entity, error)
	UpdateEntity(ctx context.Context, entity Entity, options ...GraphStoreOption) error
	DeleteEntity(ctx context.Context, id string, options ...GraphStoreOption) error

	// Relationship CRUD operations
	StoreRelationships(ctx context.Context, relationships []Relationship, options ...GraphStoreOption) error
	GetRelationships(ctx context.Context, entityID string, direction RelationshipDirection, options ...GraphSearchOption) ([]Relationship, error)
	DeleteRelationship(ctx context.Context, id string, options ...GraphStoreOption) error

	// Search operations
	Search(ctx context.Context, query string, limit int, options ...GraphSearchOption) ([]GraphSearchResult, error)
	LocalSearch(ctx context.Context, query string, entityID string, depth int, options ...GraphSearchOption) ([]GraphSearchResult, error)
	GlobalSearch(ctx context.Context, query string, communityLevel int, options ...GraphSearchOption) ([]GraphSearchResult, error)

	// Graph traversal
	TraverseFrom(ctx context.Context, entityID string, depth int, options ...GraphSearchOption) (*GraphContext, error)
	ShortestPath(ctx context.Context, sourceID, targetID string, options ...GraphSearchOption) (*GraphPath, error)

	// Entity/Relationship extraction (requires LLM)
	ExtractFromText(ctx context.Context, text string, llm LLM, options ...ExtractionOption) (*ExtractionResult, error)

	// Schema management
	ApplySchema(ctx context.Context, schema GraphSchema) error
	DiscoverSchema(ctx context.Context) (*GraphSchema, error)

	// Multi-tenancy
	SetTenant(tenant string)
	GetTenant() string

	// Lifecycle
	Close() error
}

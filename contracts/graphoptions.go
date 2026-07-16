package contracts

// GraphStoreOption represents an option for graph store operations
type GraphStoreOption func(*GraphStoreOptions)

// GraphSearchOption represents an option for graph search operations
type GraphSearchOption func(*GraphSearchOptions)

// ExtractionOption represents an option for extraction operations
type ExtractionOption func(*ExtractionOptions)

// GraphStoreOptions contains options for storing graph data
type GraphStoreOptions struct {
	// BatchSize is the number of items to store in each batch
	BatchSize int

	// GenerateEmbeddings indicates whether to generate embeddings
	GenerateEmbeddings bool

	// Tenant is the tenant name for native multi-tenancy
	Tenant string
}

// GraphSearchOptions contains options for searching graph data
type GraphSearchOptions struct {
	// MinScore is the minimum similarity score (0-1)
	MinScore float32

	// EntityTypes filters by entity types
	EntityTypes []string

	// RelationshipTypes filters by relationship types
	RelationshipTypes []string

	// MaxDepth limits traversal depth (default: 2)
	MaxDepth int

	// IncludeRelationships includes relationships in search results
	IncludeRelationships bool

	// Tenant is the tenant name for native multi-tenancy
	Tenant string

	// SearchMode specifies the search mode
	SearchMode GraphSearchMode
}

// GraphSearchMode specifies the type of search to perform
type GraphSearchMode string

const (
	// SearchModeVector uses vector similarity search
	SearchModeVector GraphSearchMode = "vector"
	// SearchModeKeyword uses keyword/BM25 search
	SearchModeKeyword GraphSearchMode = "keyword"
	// SearchModeHybrid combines vector and keyword search
	SearchModeHybrid GraphSearchMode = "hybrid"
)

// ExtractionOptions contains options for extraction operations
type ExtractionOptions struct {
	// SchemaGuided indicates whether to use schema-guided extraction
	SchemaGuided bool

	// EntityTypes limits extraction to specific entity types
	EntityTypes []string

	// RelationshipTypes limits extraction to specific relationship types
	RelationshipTypes []string

	// MinConfidence filters entities/relationships by minimum confidence
	MinConfidence float32

	// MaxEntities limits the number of entities to extract
	MaxEntities int

	// DedupThreshold is the embedding similarity threshold for deduplication
	DedupThreshold float32
}

// WithGraphBatchSize sets the batch size for store operations
func WithGraphBatchSize(size int) GraphStoreOption {
	return func(o *GraphStoreOptions) {
		o.BatchSize = size
	}
}

// WithGenerateEmbeddings sets whether to generate embeddings
func WithGenerateEmbeddings(generate bool) GraphStoreOption {
	return func(o *GraphStoreOptions) {
		o.GenerateEmbeddings = generate
	}
}

// WithGraphTenant sets the tenant for graph operations
func WithGraphTenant(tenant string) GraphStoreOption {
	return func(o *GraphStoreOptions) {
		o.Tenant = tenant
	}
}

// WithMinGraphScore sets the minimum similarity score
func WithMinGraphScore(score float32) GraphSearchOption {
	return func(o *GraphSearchOptions) {
		o.MinScore = score
	}
}

// WithEntityTypes filters search by entity types
func WithEntityTypes(types ...string) GraphSearchOption {
	return func(o *GraphSearchOptions) {
		o.EntityTypes = types
	}
}

// WithRelationshipTypes filters search by relationship types
func WithRelationshipTypes(types ...string) GraphSearchOption {
	return func(o *GraphSearchOptions) {
		o.RelationshipTypes = types
	}
}

// WithMaxDepth sets maximum traversal depth
func WithMaxDepth(depth int) GraphSearchOption {
	return func(o *GraphSearchOptions) {
		o.MaxDepth = depth
	}
}

// WithIncludeRelationships includes relationships in results
func WithIncludeRelationships(include bool) GraphSearchOption {
	return func(o *GraphSearchOptions) {
		o.IncludeRelationships = include
	}
}

// WithSearchTenant sets the tenant for search operations
func WithSearchTenant(tenant string) GraphSearchOption {
	return func(o *GraphSearchOptions) {
		o.Tenant = tenant
	}
}

// WithSearchMode sets the search mode
func WithSearchMode(mode GraphSearchMode) GraphSearchOption {
	return func(o *GraphSearchOptions) {
		o.SearchMode = mode
	}
}

// WithSchemaGuided enables schema-guided extraction
func WithSchemaGuided(guided bool) ExtractionOption {
	return func(o *ExtractionOptions) {
		o.SchemaGuided = guided
	}
}

// WithExtractionEntityTypes limits extraction to specific entity types
func WithExtractionEntityTypes(types ...string) ExtractionOption {
	return func(o *ExtractionOptions) {
		o.EntityTypes = types
	}
}

// WithExtractionRelationshipTypes limits extraction to specific relationship types
func WithExtractionRelationshipTypes(types ...string) ExtractionOption {
	return func(o *ExtractionOptions) {
		o.RelationshipTypes = types
	}
}

// WithMinConfidence sets the minimum extraction confidence
func WithMinConfidence(confidence float32) ExtractionOption {
	return func(o *ExtractionOptions) {
		o.MinConfidence = confidence
	}
}

// WithMaxEntities limits the number of extracted entities
func WithMaxEntities(max int) ExtractionOption {
	return func(o *ExtractionOptions) {
		o.MaxEntities = max
	}
}

// WithDedupThreshold sets the embedding similarity threshold for deduplication
func WithDedupThreshold(threshold float32) ExtractionOption {
	return func(o *ExtractionOptions) {
		o.DedupThreshold = threshold
	}
}

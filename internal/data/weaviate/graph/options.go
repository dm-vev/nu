package graph

import "nu/internal/contracts"

// Re-export option functions from interfaces for convenience

// Store options

// WithBatchSize sets the batch size for store operations
func WithBatchSize(size int) contracts.GraphStoreOption {
	return contracts.WithGraphBatchSize(size)
}

// WithGenerateEmbeddings sets whether to generate embeddings
func WithGenerateEmbeddings(generate bool) contracts.GraphStoreOption {
	return contracts.WithGenerateEmbeddings(generate)
}

// WithTenant sets the tenant for graph operations
func WithTenant(tenant string) contracts.GraphStoreOption {
	return contracts.WithGraphTenant(tenant)
}

// Search options

// WithMinScore sets the minimum similarity score
func WithMinScore(score float32) contracts.GraphSearchOption {
	return contracts.WithMinGraphScore(score)
}

// WithEntityTypes filters search by entity types
func WithEntityTypes(types ...string) contracts.GraphSearchOption {
	return contracts.WithEntityTypes(types...)
}

// WithRelationshipTypes filters search by relationship types
func WithRelationshipTypes(types ...string) contracts.GraphSearchOption {
	return contracts.WithRelationshipTypes(types...)
}

// WithMaxDepth sets maximum traversal depth
func WithMaxDepth(depth int) contracts.GraphSearchOption {
	return contracts.WithMaxDepth(depth)
}

// WithIncludeRelationships includes relationships in results
func WithIncludeRelationships(include bool) contracts.GraphSearchOption {
	return contracts.WithIncludeRelationships(include)
}

// WithSearchTenant sets the tenant for search operations
func WithSearchTenant(tenant string) contracts.GraphSearchOption {
	return contracts.WithSearchTenant(tenant)
}

// WithMode sets the search mode (vector, keyword, hybrid)
func WithSearchMode(mode SearchMode) contracts.GraphSearchOption {
	return contracts.WithSearchMode(mode)
}

// Extraction options

// WithSchemaGuided enables schema-guided extraction
func WithSchemaGuided(guided bool) contracts.ExtractionOption {
	return contracts.WithSchemaGuided(guided)
}

// WithExtractionEntityTypes limits extraction to specific entity types
func WithExtractionEntityTypes(types ...string) contracts.ExtractionOption {
	return contracts.WithExtractionEntityTypes(types...)
}

// WithExtractionRelationshipTypes limits extraction to specific relationship types
func WithExtractionRelationshipTypes(types ...string) contracts.ExtractionOption {
	return contracts.WithExtractionRelationshipTypes(types...)
}

// WithMinConfidence sets the minimum extraction confidence
func WithMinConfidence(confidence float32) contracts.ExtractionOption {
	return contracts.WithMinConfidence(confidence)
}

// WithMaxEntities limits the number of extracted entities
func WithMaxEntities(max int) contracts.ExtractionOption {
	return contracts.WithMaxEntities(max)
}

// WithDedupThreshold sets the embedding similarity threshold for deduplication
func WithDedupThreshold(threshold float32) contracts.ExtractionOption {
	return contracts.WithDedupThreshold(threshold)
}

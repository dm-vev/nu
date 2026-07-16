package graph

import "nu/internal/contracts"

// applyStoreOptions applies GraphStoreOptions and returns the effective options.
func applyStoreOptions(opts []contracts.GraphStoreOption) *contracts.GraphStoreOptions {
	options := &contracts.GraphStoreOptions{
		BatchSize:          100, // Default batch size
		GenerateEmbeddings: true,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// applySearchOptions applies GraphSearchOptions and returns the effective options.
func applySearchOptions(opts []contracts.GraphSearchOption) *contracts.GraphSearchOptions {
	options := &contracts.GraphSearchOptions{
		MinScore:             0.0,
		MaxDepth:             2,
		IncludeRelationships: false,
		SearchMode:           contracts.SearchModeHybrid,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// applyExtractionOptions applies ExtractionOptions and returns the effective options.
func applyExtractionOptions(opts []contracts.ExtractionOption) *contracts.ExtractionOptions {
	options := &contracts.ExtractionOptions{
		SchemaGuided:   false,
		MinConfidence:  0.5,
		MaxEntities:    50,
		DedupThreshold: 0.92,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

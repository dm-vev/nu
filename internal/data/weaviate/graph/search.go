package graph

import (
	"context"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// Search performs a search on the knowledge graph.
// It supports vector, keyword, and hybrid search modes.
func (s *Store) Search(ctx context.Context, query string, limit int, opts ...contracts.GraphSearchOption) ([]contracts.GraphSearchResult, error) {
	if query == "" {
		return nil, ErrEmptyQuery
	}

	if limit <= 0 {
		limit = 10
	}

	options := applySearchOptions(opts)

	// Get tenant from options, context, or store default
	tenant := options.Tenant
	if tenant == "" {
		if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
			tenant = orgID
		} else {
			tenant = s.tenant
		}
	}

	// Build the query based on search mode
	var results []contracts.GraphSearchResult
	var err error

	switch options.SearchMode {
	case contracts.SearchModeKeyword:
		results, err = s.keywordSearch(ctx, query, limit, tenant, options)
	case contracts.SearchModeVector:
		results, err = s.vectorSearch(ctx, query, limit, tenant, options)
	default: // Hybrid
		results, err = s.hybridSearch(ctx, query, limit, tenant, options)
	}

	if err != nil {
		return nil, err
	}

	// Filter by minimum score
	if options.MinScore > 0 {
		filtered := make([]contracts.GraphSearchResult, 0, len(results))
		for _, r := range results {
			if r.Score >= options.MinScore {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	s.logger.Debug(ctx, "Search completed", map[string]interface{}{
		"query":  query,
		"mode":   options.SearchMode,
		"count":  len(results),
		"tenant": tenant,
	})

	return results, nil
}

// LocalSearch performs a search starting from a specific entity and traversing the graph.
func (s *Store) LocalSearch(ctx context.Context, query string, entityID string, depth int, opts ...contracts.GraphSearchOption) ([]contracts.GraphSearchResult, error) {
	if query == "" {
		return nil, ErrEmptyQuery
	}

	if depth < 0 {
		return nil, ErrInvalidDepth
	}
	if depth == 0 {
		depth = 2 // Default depth
	}

	options := applySearchOptions(opts)

	// First, perform a regular search to find relevant entities
	searchResults, err := s.Search(ctx, query, 10, opts...)
	if err != nil {
		return nil, err
	}

	// If an entity ID is provided, get context from that entity
	if entityID != "" {
		graphContext, err := s.TraverseFrom(ctx, entityID, depth, opts...)
		if err != nil && err != ErrEntityNotFound {
			return nil, err
		}

		if graphContext != nil {
			// Enrich search results with context
			for i := range searchResults {
				searchResults[i].Context = graphContext.Entities
				if options.IncludeRelationships {
					searchResults[i].Path = graphContext.Relationships
				}
			}
		}
	} else if len(searchResults) > 0 {
		// Use the top result as the starting point
		topEntityID := searchResults[0].Entity.ID
		graphContext, err := s.TraverseFrom(ctx, topEntityID, depth, opts...)
		if err != nil && err != ErrEntityNotFound {
			s.logger.Warn(ctx, "Failed to get context for top result", map[string]interface{}{
				"entityId": topEntityID,
				"error":    err.Error(),
			})
		}

		if graphContext != nil {
			searchResults[0].Context = graphContext.Entities
			if options.IncludeRelationships {
				searchResults[0].Path = graphContext.Relationships
			}
		}
	}

	s.logger.Debug(ctx, "Local search completed", map[string]interface{}{
		"query":    query,
		"entityId": entityID,
		"depth":    depth,
		"count":    len(searchResults),
	})

	return searchResults, nil
}

// GlobalSearch performs a community-based search across the knowledge graph.
// It groups entities by type and searches across communities.
func (s *Store) GlobalSearch(ctx context.Context, query string, communityLevel int, opts ...contracts.GraphSearchOption) ([]contracts.GraphSearchResult, error) {
	if query == "" {
		return nil, ErrEmptyQuery
	}

	options := applySearchOptions(opts)

	// Get tenant from options, context, or store default
	tenant := options.Tenant
	if tenant == "" {
		if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
			tenant = orgID
		} else {
			tenant = s.tenant
		}
	}

	// Get all entity types (communities)
	entityTypes, err := s.discoverEntityTypes(ctx)
	if err != nil {
		s.logger.Warn(ctx, "Failed to discover entity types, using search without community grouping", map[string]interface{}{
			"error": err.Error(),
		})
		return s.Search(ctx, query, 20, opts...)
	}

	// If specific entity types are requested, filter to those
	if len(options.EntityTypes) > 0 {
		filtered := []string{}
		typeSet := make(map[string]bool)
		for _, t := range options.EntityTypes {
			typeSet[t] = true
		}
		for _, t := range entityTypes {
			if typeSet[t] {
				filtered = append(filtered, t)
			}
		}
		entityTypes = filtered
	}

	// Search within each entity type (community)
	allResults := []contracts.GraphSearchResult{}
	resultsPerCommunity := 5

	for _, entityType := range entityTypes {
		// Search within this entity type
		typeOpts := append(opts, contracts.WithEntityTypes(entityType))
		results, err := s.Search(ctx, query, resultsPerCommunity, typeOpts...)
		if err != nil {
			s.logger.Warn(ctx, "Failed to search in entity type", map[string]interface{}{
				"entityType": entityType,
				"error":      err.Error(),
			})
			continue
		}

		// Add community ID to results
		for i := range results {
			results[i].CommunityID = entityType
		}

		allResults = append(allResults, results...)
	}

	// Sort by score (already sorted within each community, but need global sort)
	// Simple bubble sort since we expect small result sets
	for i := 0; i < len(allResults); i++ {
		for j := i + 1; j < len(allResults); j++ {
			if allResults[j].Score > allResults[i].Score {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}

	s.logger.Debug(ctx, "Global search completed", map[string]interface{}{
		"query":       query,
		"communities": len(entityTypes),
		"count":       len(allResults),
		"tenant":      tenant,
	})

	return allResults, nil
}

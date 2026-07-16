package graph

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/weaviate/graph/search"
)

// vectorSearch performs a pure vector similarity search.
func (s *Store) vectorSearch(ctx context.Context, query string, limit int, tenant string, options *contracts.GraphSearchOptions) ([]contracts.GraphSearchResult, error) {
	if s.embedder == nil {
		return nil, ErrNoEmbedder
	}

	className := s.getEntityClassName()

	// Generate query embedding
	queryVector, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Build filter
	filter := search.BuildFilter(tenant, options.EntityTypes)

	// Build query
	nearVector := s.client.GraphQL().NearVectorArgBuilder().
		WithVector(queryVector)

	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "entityId"},
			graphql.Field{Name: "name"},
			graphql.Field{Name: "entityType"},
			graphql.Field{Name: "description"},
			graphql.Field{Name: "properties"},
			graphql.Field{Name: "orgId"},
			graphql.Field{Name: "createdAt"},
			graphql.Field{Name: "updatedAt"},
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "id"},
				{Name: "certainty"},
				{Name: "vector"},
			}},
		).
		WithNearVector(nearVector).
		WithLimit(limit)

	if filter != nil {
		queryBuilder = queryBuilder.WithWhere(filter)
	}

	result, err := queryBuilder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute vector search: %w", err)
	}

	return search.ParseResults(result, className), nil
}

// keywordSearch performs a BM25 keyword search.
func (s *Store) keywordSearch(ctx context.Context, query string, limit int, tenant string, options *contracts.GraphSearchOptions) ([]contracts.GraphSearchResult, error) {
	className := s.getEntityClassName()

	// Build filter
	filter := search.BuildFilter(tenant, options.EntityTypes)

	// Build BM25 query
	bm25 := s.client.GraphQL().Bm25ArgBuilder().
		WithQuery(query).
		WithProperties("name", "description", "entityType")

	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "entityId"},
			graphql.Field{Name: "name"},
			graphql.Field{Name: "entityType"},
			graphql.Field{Name: "description"},
			graphql.Field{Name: "properties"},
			graphql.Field{Name: "orgId"},
			graphql.Field{Name: "createdAt"},
			graphql.Field{Name: "updatedAt"},
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "id"},
				{Name: "score"},
			}},
		).
		WithBM25(bm25).
		WithLimit(limit)

	if filter != nil {
		queryBuilder = queryBuilder.WithWhere(filter)
	}

	result, err := queryBuilder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute keyword search: %w", err)
	}

	return search.ParseResults(result, className), nil
}

// hybridSearch performs a hybrid search combining vector and keyword search.
func (s *Store) hybridSearch(ctx context.Context, query string, limit int, tenant string, options *contracts.GraphSearchOptions) ([]contracts.GraphSearchResult, error) {
	className := s.getEntityClassName()

	s.logger.Info(ctx, "hybridSearch starting", map[string]interface{}{
		"query":       query,
		"limit":       limit,
		"tenant":      tenant,
		"entityTypes": options.EntityTypes,
		"className":   className,
		"hasEmbedder": s.embedder != nil,
	})

	// If no embedder is available, fall back to keyword search
	// Hybrid search without vectors doesn't work properly in Weaviate
	if s.embedder == nil {
		s.logger.Info(ctx, "No embedder configured, falling back to keyword search", nil)
		return s.keywordSearch(ctx, query, limit, tenant, options)
	}

	// Generate query vector
	queryVector, err := s.embedder.Embed(ctx, query)
	if err != nil {
		s.logger.Warn(ctx, "Failed to generate embedding for hybrid search, falling back to keyword", map[string]interface{}{
			"error": err.Error(),
		})
		return s.keywordSearch(ctx, query, limit, tenant, options)
	}

	// Build filter
	filter := search.BuildFilter(tenant, options.EntityTypes)
	s.logger.Info(ctx, "hybridSearch filter built", map[string]interface{}{
		"hasFilter": filter != nil,
	})

	// Build hybrid query with vector
	hybridBuilder := s.client.GraphQL().HybridArgumentBuilder().
		WithQuery(query).
		WithAlpha(0.5). // Equal weight to vector and keyword
		WithVector(queryVector)

	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "entityId"},
			graphql.Field{Name: "name"},
			graphql.Field{Name: "entityType"},
			graphql.Field{Name: "description"},
			graphql.Field{Name: "properties"},
			graphql.Field{Name: "orgId"},
			graphql.Field{Name: "createdAt"},
			graphql.Field{Name: "updatedAt"},
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "id"},
				{Name: "score"},
				{Name: "vector"},
			}},
		).
		WithHybrid(hybridBuilder).
		WithLimit(limit)

	if filter != nil {
		queryBuilder = queryBuilder.WithWhere(filter)
	}

	result, err := queryBuilder.Do(ctx)
	if err != nil {
		s.logger.Error(ctx, "hybridSearch failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute hybrid search: %w", err)
	}

	// Log raw result
	if result.Data != nil {
		if getData, ok := result.Data["Get"].(map[string]interface{}); ok {
			if classData, ok := getData[className].([]interface{}); ok {
				s.logger.Info(ctx, "hybridSearch raw result", map[string]interface{}{
					"resultCount": len(classData),
				})
			} else {
				s.logger.Info(ctx, "hybridSearch no class data found", map[string]interface{}{
					"className": className,
					"getData":   getData,
				})
			}
		} else {
			s.logger.Info(ctx, "hybridSearch no Get data", map[string]interface{}{
				"data": result.Data,
			})
		}
	} else {
		s.logger.Info(ctx, "hybridSearch result.Data is nil", nil)
	}

	return search.ParseResults(result, className), nil
}

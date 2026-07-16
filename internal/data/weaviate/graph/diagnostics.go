package graph

import (
	"context"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"

	"nu/internal/contracts"
	"nu/internal/data/weaviate/graph/entity"
)

// CountAllEntities is a diagnostic method that counts all entities without any filter.
// This is useful for debugging to verify data is actually stored.
func (s *Store) CountAllEntities(ctx context.Context) (int, error) {
	className := s.getEntityClassName()

	// Use aggregate query to count all entities
	result, err := s.client.GraphQL().Aggregate().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "meta", Fields: []graphql.Field{{Name: "count"}}},
		).
		Do(ctx)

	if err != nil {
		s.logger.Error(ctx, "CountAllEntities failed", map[string]interface{}{
			"error": err.Error(),
		})
		return 0, err
	}

	s.logger.Info(ctx, "CountAllEntities raw result", map[string]interface{}{
		"data": result.Data,
	})

	// Parse the aggregate response
	if result.Data != nil {
		if aggData, ok := result.Data["Aggregate"].(map[string]interface{}); ok {
			if classData, ok := aggData[className].([]interface{}); ok {
				if len(classData) > 0 {
					if item, ok := classData[0].(map[string]interface{}); ok {
						if meta, ok := item["meta"].(map[string]interface{}); ok {
							if count, ok := meta["count"].(float64); ok {
								return int(count), nil
							}
						}
					}
				}
			}
		}
	}

	return 0, nil
}

// ListAllEntities is a diagnostic method that lists all entities without any filter.
func (s *Store) ListAllEntities(ctx context.Context, limit int) ([]contracts.Entity, error) {
	className := s.getEntityClassName()

	if limit <= 0 {
		limit = 100
	}

	// Query all entities without filter
	result, err := s.client.GraphQL().Get().
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
		).
		WithLimit(limit).
		Do(ctx)

	if err != nil {
		s.logger.Error(ctx, "ListAllEntities failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Log raw result
	if result.Data != nil {
		if getData, ok := result.Data["Get"].(map[string]interface{}); ok {
			if classData, ok := getData[className].([]interface{}); ok {
				s.logger.Info(ctx, "ListAllEntities raw result", map[string]interface{}{
					"count": len(classData),
				})
			} else {
				s.logger.Warn(ctx, "ListAllEntities no class data", map[string]interface{}{
					"className": className,
					"getData":   getData,
				})
			}
		}
	}

	return entity.ParseResults(result, className), nil
}

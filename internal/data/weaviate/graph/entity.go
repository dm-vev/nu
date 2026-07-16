package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"

	"nu/internal/contracts"
	"nu/internal/data/weaviate/graph/entity"
	"nu/internal/multitenancy"
)

// StoreEntities stores multiple entities in the knowledge graph.
func (s *Store) StoreEntities(ctx context.Context, entities []contracts.Entity, opts ...contracts.GraphStoreOption) error {
	if len(entities) == 0 {
		return nil
	}

	options := applyStoreOptions(opts)
	className := s.getEntityClassName()

	// Get tenant from options, context, or store default
	tenant := options.Tenant
	if tenant == "" {
		if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
			tenant = orgID
		} else {
			tenant = s.tenant
		}
	}

	// Create batch
	batch := s.client.Batch().ObjectsBatcher()
	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	batchCount := 0

	for i := range entities {
		entity := &entities[i]

		// Validate entity
		if entity.ID == "" {
			return fmt.Errorf("%w: entity at index %d", ErrInvalidEntityID, i)
		}
		if entity.Name == "" {
			return fmt.Errorf("%w: entity %s", ErrMissingEntityName, entity.ID)
		}
		if entity.Type == "" {
			return fmt.Errorf("%w: entity %s", ErrMissingEntityType, entity.ID)
		}

		// Generate embedding if needed
		var vector []float32
		if options.GenerateEmbeddings && s.embedder != nil {
			// Use description for embedding, fall back to name if description is empty
			textToEmbed := entity.Description
			if textToEmbed == "" {
				textToEmbed = entity.Name
			}

			embedding, err := s.embedder.Embed(ctx, textToEmbed)
			if err != nil {
				return fmt.Errorf("failed to generate embedding for entity %s: %w", entity.ID, err)
			}
			vector = embedding
		} else if len(entity.Embedding) > 0 {
			vector = entity.Embedding
		}

		// Serialize properties to JSON
		propertiesJSON := ""
		if entity.Properties != nil {
			data, err := json.Marshal(entity.Properties)
			if err != nil {
				return fmt.Errorf("failed to serialize properties for entity %s: %w", entity.ID, err)
			}
			propertiesJSON = string(data)
		}

		// Set timestamps
		now := time.Now()
		if entity.CreatedAt.IsZero() {
			entity.CreatedAt = now
		}
		if entity.UpdatedAt.IsZero() {
			entity.UpdatedAt = now
		}

		// Create object properties
		props := map[string]interface{}{
			"entityId":    entity.ID,
			"name":        entity.Name,
			"entityType":  entity.Type,
			"description": entity.Description,
			"properties":  propertiesJSON,
			"orgId":       tenant,
			"createdAt":   entity.CreatedAt.Format(time.RFC3339),
			"updatedAt":   entity.UpdatedAt.Format(time.RFC3339),
		}

		// Create Weaviate object
		obj := &models.Object{
			Class:      className,
			Properties: props,
		}

		if len(vector) > 0 {
			obj.Vector = vector
		}

		batch.WithObjects(obj)
		batchCount++

		// Execute batch if size reached
		if batchCount >= batchSize {
			result, err := batch.Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to store entity batch: %w", err)
			}

			// Check for individual object errors
			for _, res := range result {
				if res.Result != nil && res.Result.Errors != nil {
					for _, objErr := range res.Result.Errors.Error {
						s.logger.Error(ctx, "Failed to store entity", map[string]interface{}{
							"error": objErr.Message,
						})
					}
				}
			}

			// Reset batch
			batch = s.client.Batch().ObjectsBatcher()
			batchCount = 0
		}
	}

	// Store remaining entities
	if batchCount > 0 {
		result, err := batch.Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to store final entity batch: %w", err)
		}

		s.logger.Info(ctx, "Batch result received", map[string]interface{}{
			"resultCount": len(result),
			"batchCount":  batchCount,
		})

		// Check for individual object errors
		successCount := 0
		for _, res := range result {
			if res.Result != nil && res.Result.Errors != nil {
				for _, objErr := range res.Result.Errors.Error {
					s.logger.Error(ctx, "Failed to store entity", map[string]interface{}{
						"error": objErr.Message,
					})
				}
			} else {
				successCount++
			}
		}
		s.logger.Info(ctx, "Batch store completed", map[string]interface{}{
			"successCount": successCount,
			"totalCount":   len(result),
		})
	}

	s.logger.Info(ctx, "Stored entities", map[string]interface{}{
		"count":  len(entities),
		"tenant": tenant,
	})

	return nil
}

// GetEntity retrieves an entity by its ID.
func (s *Store) GetEntity(ctx context.Context, id string, opts ...contracts.GraphStoreOption) (*contracts.Entity, error) {
	if id == "" {
		return nil, ErrInvalidEntityID
	}

	options := applyStoreOptions(opts)
	className := s.getEntityClassName()

	// Get tenant from options, context, or store default
	tenant := options.Tenant
	if tenant == "" {
		if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
			tenant = orgID
		} else {
			tenant = s.tenant
		}
	}

	s.logger.Info(ctx, "GetEntity starting", map[string]interface{}{
		"entityId":  id,
		"tenant":    tenant,
		"className": className,
	})

	// Build filter for entityId and optionally orgId
	filter := entity.BuildIDFilter(id, tenant)

	// Query for the entity
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
				{Name: "vector"},
			}},
		).
		WithLimit(1)

	if filter != nil {
		queryBuilder = queryBuilder.WithWhere(filter)
	}

	result, err := queryBuilder.Do(ctx)

	if err != nil {
		s.logger.Error(ctx, "GetEntity query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to query entity: %w", err)
	}

	// Log raw result
	if result.Data != nil {
		if getData, ok := result.Data["Get"].(map[string]interface{}); ok {
			if classData, ok := getData[className].([]interface{}); ok {
				s.logger.Info(ctx, "GetEntity raw result", map[string]interface{}{
					"resultCount": len(classData),
				})
			} else {
				s.logger.Info(ctx, "GetEntity no class data found", map[string]interface{}{
					"className": className,
					"getData":   getData,
				})
			}
		}
	}

	// Parse result
	entities := entity.ParseResults(result, className)
	if len(entities) == 0 {
		s.logger.Warn(ctx, "GetEntity returned 0 entities", map[string]interface{}{
			"entityId": id,
			"tenant":   tenant,
		})
		return nil, ErrEntityNotFound
	}

	s.logger.Info(ctx, "GetEntity found entity", map[string]interface{}{
		"entityId": entities[0].ID,
		"name":     entities[0].Name,
		"orgId":    entities[0].OrgID,
	})

	return &entities[0], nil
}

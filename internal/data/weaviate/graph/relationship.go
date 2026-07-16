package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/weaviate/graph/relationship"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// StoreRelationships stores multiple relationships in the knowledge graph.
func (s *Store) StoreRelationships(ctx context.Context, relationships []contracts.Relationship, opts ...contracts.GraphStoreOption) error {
	if len(relationships) == 0 {
		return nil
	}

	options := applyStoreOptions(opts)
	className := s.getRelationshipClassName()

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

	for i := range relationships {
		rel := &relationships[i]

		// Validate relationship
		if rel.ID == "" {
			return fmt.Errorf("%w: relationship at index %d", ErrInvalidRelationshipID, i)
		}
		if rel.SourceID == "" {
			return fmt.Errorf("%w: relationship %s", ErrMissingSourceID, rel.ID)
		}
		if rel.TargetID == "" {
			return fmt.Errorf("%w: relationship %s", ErrMissingTargetID, rel.ID)
		}
		if rel.Type == "" {
			return fmt.Errorf("%w: relationship %s", ErrMissingRelationshipType, rel.ID)
		}
		if rel.Strength < 0 || rel.Strength > 1 {
			return fmt.Errorf("%w: relationship %s has strength %.2f", ErrInvalidStrength, rel.ID, rel.Strength)
		}

		// Default strength to 1.0 if not set
		strength := rel.Strength
		if strength == 0 {
			strength = 1.0
		}

		// Serialize properties to JSON
		propertiesJSON := ""
		if rel.Properties != nil {
			data, err := json.Marshal(rel.Properties)
			if err != nil {
				return fmt.Errorf("failed to serialize properties for relationship %s: %w", rel.ID, err)
			}
			propertiesJSON = string(data)
		}

		// Set timestamp
		if rel.CreatedAt.IsZero() {
			rel.CreatedAt = time.Now()
		}

		// Create object properties
		props := map[string]interface{}{
			"relationshipId":   rel.ID,
			"sourceId":         rel.SourceID,
			"targetId":         rel.TargetID,
			"relationshipType": rel.Type,
			"description":      rel.Description,
			"strength":         strength,
			"properties":       propertiesJSON,
			"orgId":            tenant,
			"createdAt":        rel.CreatedAt.Format(time.RFC3339),
		}

		// Create Weaviate object
		obj := &models.Object{
			Class:      className,
			Properties: props,
		}

		batch.WithObjects(obj)
		batchCount++

		// Execute batch if size reached
		if batchCount >= batchSize {
			result, err := batch.Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to store relationship batch: %w", err)
			}

			// Check for individual object errors
			for _, res := range result {
				if res.Result != nil && res.Result.Errors != nil {
					for _, objErr := range res.Result.Errors.Error {
						s.logger.Error(ctx, "Failed to store relationship", map[string]interface{}{
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

	// Store remaining relationships
	if batchCount > 0 {
		result, err := batch.Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to store final relationship batch: %w", err)
		}

		// Check for individual object errors
		for _, res := range result {
			if res.Result != nil && res.Result.Errors != nil {
				for _, objErr := range res.Result.Errors.Error {
					s.logger.Error(ctx, "Failed to store relationship", map[string]interface{}{
						"error": objErr.Message,
					})
				}
			}
		}
	}

	s.logger.Info(ctx, "Stored relationships", map[string]interface{}{
		"count":  len(relationships),
		"tenant": tenant,
	})

	return nil
}

// GetRelationships retrieves relationships for an entity based on direction.
func (s *Store) GetRelationships(ctx context.Context, entityID string, direction contracts.RelationshipDirection, opts ...contracts.GraphSearchOption) ([]contracts.Relationship, error) {
	if entityID == "" {
		return nil, ErrInvalidEntityID
	}

	options := applySearchOptions(opts)
	className := s.getRelationshipClassName()

	// Get tenant from options, context, or store default
	tenant := options.Tenant
	if tenant == "" {
		if orgID, err := multitenancy.GetOrgID(ctx); err == nil {
			tenant = orgID
		} else {
			tenant = s.tenant
		}
	}

	// Build filter based on direction
	filter := relationship.BuildDirectionFilter(entityID, direction, tenant, options.RelationshipTypes)

	// Query for relationships
	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "relationshipId"},
			graphql.Field{Name: "sourceId"},
			graphql.Field{Name: "targetId"},
			graphql.Field{Name: "relationshipType"},
			graphql.Field{Name: "description"},
			graphql.Field{Name: "strength"},
			graphql.Field{Name: "properties"},
			graphql.Field{Name: "orgId"},
			graphql.Field{Name: "createdAt"},
		)

	if filter != nil {
		queryBuilder = queryBuilder.WithWhere(filter)
	}

	result, err := queryBuilder.Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}

	relationships := relationship.ParseResults(result, className)

	s.logger.Debug(ctx, "Retrieved relationships", map[string]interface{}{
		"entityId":  entityID,
		"direction": direction,
		"count":     len(relationships),
	})

	return relationships, nil
}

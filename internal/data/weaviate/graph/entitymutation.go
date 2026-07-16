package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/weaviate/graph/entity"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// UpdateEntity updates an existing entity.
func (s *Store) UpdateEntity(ctx context.Context, item contracts.Entity, opts ...contracts.GraphStoreOption) error {
	if item.ID == "" {
		return ErrInvalidEntityID
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

	// First, find the Weaviate UUID for this entity
	filter := entity.BuildIDFilter(item.ID, tenant)

	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "id"},
			}},
		).
		WithLimit(1)

	if filter != nil {
		queryBuilder = queryBuilder.WithWhere(filter)
	}

	result, err := queryBuilder.Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to find entity: %w", err)
	}

	// Extract UUID
	uuid := entity.ExtractUUID(result, className)
	if uuid == "" {
		return ErrEntityNotFound
	}

	// Generate new embedding if needed
	var vector []float32
	if options.GenerateEmbeddings && s.embedder != nil {
		textToEmbed := item.Description
		if textToEmbed == "" {
			textToEmbed = item.Name
		}

		embedding, err := s.embedder.Embed(ctx, textToEmbed)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		vector = embedding
	} else if len(item.Embedding) > 0 {
		vector = item.Embedding
	}

	// Serialize properties
	propertiesJSON := ""
	if item.Properties != nil {
		data, err := json.Marshal(item.Properties)
		if err != nil {
			return fmt.Errorf("failed to serialize properties: %w", err)
		}
		propertiesJSON = string(data)
	}

	// Update timestamp
	item.UpdatedAt = time.Now()

	// Update the entity
	props := map[string]interface{}{
		"entityId":    item.ID,
		"name":        item.Name,
		"entityType":  item.Type,
		"description": item.Description,
		"properties":  propertiesJSON,
		"orgId":       tenant,
		"updatedAt":   item.UpdatedAt.Format(time.RFC3339),
	}

	err = s.client.Data().Updater().
		WithClassName(className).
		WithID(uuid).
		WithProperties(props).
		WithVector(vector).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	s.logger.Info(ctx, "Updated entity", map[string]interface{}{
		"entityId": item.ID,
		"tenant":   tenant,
	})

	return nil
}

// DeleteEntity deletes an entity by its ID.
func (s *Store) DeleteEntity(ctx context.Context, id string, opts ...contracts.GraphStoreOption) error {
	if id == "" {
		return ErrInvalidEntityID
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

	// Find the Weaviate UUID for this entity
	filter := entity.BuildIDFilter(id, tenant)

	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "_additional", Fields: []graphql.Field{
				{Name: "id"},
			}},
		).
		WithLimit(1)

	if filter != nil {
		queryBuilder = queryBuilder.WithWhere(filter)
	}

	result, err := queryBuilder.Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to find entity: %w", err)
	}

	// Extract UUID
	uuid := entity.ExtractUUID(result, className)
	if uuid == "" {
		return ErrEntityNotFound
	}

	// Delete the entity
	err = s.client.Data().Deleter().
		WithClassName(className).
		WithID(uuid).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	s.logger.Info(ctx, "Deleted entity", map[string]interface{}{
		"entityId": id,
		"tenant":   tenant,
	})

	return nil
}

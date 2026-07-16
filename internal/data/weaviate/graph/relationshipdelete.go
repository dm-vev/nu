package graph

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"

	"nu/internal/contracts"
	"nu/internal/data/weaviate/graph/relationship"
	"nu/internal/multitenancy"
)

// DeleteRelationship deletes a relationship by its ID.
func (s *Store) DeleteRelationship(ctx context.Context, id string, opts ...contracts.GraphStoreOption) error {
	if id == "" {
		return ErrInvalidRelationshipID
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

	// Find the Weaviate UUID for this relationship
	filter := relationship.BuildIDFilter(id, tenant)

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
		return fmt.Errorf("failed to find relationship: %w", err)
	}

	// Extract UUID
	uuid := relationship.ExtractUUID(result, className)
	if uuid == "" {
		return ErrRelationshipNotFound
	}

	// Delete the relationship
	err = s.client.Data().Deleter().
		WithClassName(className).
		WithID(uuid).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	s.logger.Info(ctx, "Deleted relationship", map[string]interface{}{
		"relationshipId": id,
		"tenant":         tenant,
	})

	return nil
}

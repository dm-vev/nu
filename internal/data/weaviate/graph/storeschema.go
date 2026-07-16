package graph

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"

	"nu/internal/contracts"
)

// ApplySchema stores the schema definition for use in extraction and validation.
func (s *Store) ApplySchema(ctx context.Context, schema contracts.GraphSchema) error {
	s.schema = &schema
	s.logger.Info(ctx, "Applied graph schema", map[string]interface{}{
		"entity_types":       len(schema.EntityTypes),
		"relationship_types": len(schema.RelationshipTypes),
	})
	return nil
}

// DiscoverSchema infers schema from existing data in the graph.
func (s *Store) DiscoverSchema(ctx context.Context) (*contracts.GraphSchema, error) {
	// If we have a stored schema, return it
	if s.schema != nil {
		return s.schema, nil
	}

	schema := &contracts.GraphSchema{
		EntityTypes:       []contracts.EntityTypeSchema{},
		RelationshipTypes: []contracts.RelationshipTypeSchema{},
	}

	// Discover entity types by querying unique entityType values
	entityTypes, err := s.discoverEntityTypes(ctx)
	if err != nil {
		s.logger.Warn(ctx, "Failed to discover entity types", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		for _, et := range entityTypes {
			schema.EntityTypes = append(schema.EntityTypes, contracts.EntityTypeSchema{
				Name:        et,
				Description: fmt.Sprintf("Discovered entity type: %s", et),
			})
		}
	}

	// Discover relationship types
	relTypes, err := s.discoverRelationshipTypes(ctx)
	if err != nil {
		s.logger.Warn(ctx, "Failed to discover relationship types", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		for _, rt := range relTypes {
			schema.RelationshipTypes = append(schema.RelationshipTypes, contracts.RelationshipTypeSchema{
				Name:        rt,
				Description: fmt.Sprintf("Discovered relationship type: %s", rt),
			})
		}
	}

	return schema, nil
}

// discoverEntityTypes queries for unique entity types in the collection.
func (s *Store) discoverEntityTypes(ctx context.Context) ([]string, error) {
	className := s.getEntityClassName()

	// Use aggregate query to get unique entity types
	result, err := s.client.GraphQL().Aggregate().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "entityType", Fields: []graphql.Field{{Name: "count"}}},
		).
		WithGroupBy("entityType").
		Do(ctx)

	if err != nil {
		return nil, err
	}

	types := []string{}

	// Parse the aggregate response
	if result.Data != nil {
		if aggData, ok := result.Data["Aggregate"].(map[string]interface{}); ok {
			if classData, ok := aggData[className].([]interface{}); ok {
				for _, item := range classData {
					if group, ok := item.(map[string]interface{}); ok {
						if groupedBy, ok := group["groupedBy"].(map[string]interface{}); ok {
							if value, ok := groupedBy["value"].(string); ok {
								types = append(types, value)
							}
						}
					}
				}
			}
		}
	}

	return types, nil
}

// discoverRelationshipTypes queries for unique relationship types in the collection.
func (s *Store) discoverRelationshipTypes(ctx context.Context) ([]string, error) {
	className := s.getRelationshipClassName()

	// Use aggregate query to get unique relationship types
	result, err := s.client.GraphQL().Aggregate().
		WithClassName(className).
		WithFields(
			graphql.Field{Name: "relationshipType", Fields: []graphql.Field{{Name: "count"}}},
		).
		WithGroupBy("relationshipType").
		Do(ctx)

	if err != nil {
		return nil, err
	}

	types := []string{}

	// Parse the aggregate response
	if result.Data != nil {
		if aggData, ok := result.Data["Aggregate"].(map[string]interface{}); ok {
			if classData, ok := aggData[className].([]interface{}); ok {
				for _, item := range classData {
					if group, ok := item.(map[string]interface{}); ok {
						if groupedBy, ok := group["groupedBy"].(map[string]interface{}); ok {
							if value, ok := groupedBy["value"].(string); ok {
								types = append(types, value)
							}
						}
					}
				}
			}
		}
	}

	return types, nil
}

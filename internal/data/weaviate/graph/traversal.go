package graph

import (
	"context"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// TraverseFrom performs a breadth-first traversal from a starting entity.
// Note: Weaviate doesn't support native graph traversal, so this requires
// multiple queries (one per hop level).
func (s *Store) TraverseFrom(ctx context.Context, entityID string, depth int, opts ...contracts.GraphSearchOption) (*contracts.GraphContext, error) {
	if entityID == "" {
		return nil, ErrInvalidEntityID
	}

	if depth < 0 {
		return nil, ErrInvalidDepth
	}
	if depth == 0 {
		depth = 2 // Default depth
	}
	if depth > 5 {
		return nil, ErrMaxDepthExceeded
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

	result := &contracts.GraphContext{
		Depth:         depth,
		Entities:      []contracts.Entity{},
		Relationships: []contracts.Relationship{},
	}

	visited := make(map[string]bool)
	currentLevel := []string{entityID}

	for d := 0; d <= depth && len(currentLevel) > 0; d++ {
		nextLevel := []string{}

		for _, id := range currentLevel {
			if visited[id] {
				continue
			}
			visited[id] = true

			// Get the entity
			storeOpts := []contracts.GraphStoreOption{}
			if tenant != "" {
				storeOpts = append(storeOpts, contracts.WithGraphTenant(tenant))
			}

			entity, err := s.GetEntity(ctx, id, storeOpts...)
			if err != nil {
				s.logger.Debug(ctx, "Entity not found during traversal", map[string]interface{}{
					"entityId": id,
					"error":    err.Error(),
				})
				continue
			}

			// Set central entity for the first entity
			if d == 0 {
				result.CentralEntity = *entity
			}
			result.Entities = append(result.Entities, *entity)

			// Get outgoing relationships (we traverse in both directions for completeness)
			searchOpts := []contracts.GraphSearchOption{}
			if tenant != "" {
				searchOpts = append(searchOpts, contracts.WithSearchTenant(tenant))
			}
			if len(options.RelationshipTypes) > 0 {
				searchOpts = append(searchOpts, contracts.WithRelationshipTypes(options.RelationshipTypes...))
			}

			outgoing, err := s.GetRelationships(ctx, id, contracts.DirectionOutgoing, searchOpts...)
			if err != nil {
				s.logger.Debug(ctx, "Failed to get outgoing relationships", map[string]interface{}{
					"entityId": id,
					"error":    err.Error(),
				})
			} else {
				for _, rel := range outgoing {
					result.Relationships = append(result.Relationships, rel)
					if !visited[rel.TargetID] {
						nextLevel = append(nextLevel, rel.TargetID)
					}
				}
			}

			// Also get incoming relationships for complete graph context
			incoming, err := s.GetRelationships(ctx, id, contracts.DirectionIncoming, searchOpts...)
			if err != nil {
				s.logger.Debug(ctx, "Failed to get incoming relationships", map[string]interface{}{
					"entityId": id,
					"error":    err.Error(),
				})
			} else {
				for _, rel := range incoming {
					// Avoid duplicates
					isDuplicate := false
					for _, existing := range result.Relationships {
						if existing.ID == rel.ID {
							isDuplicate = true
							break
						}
					}
					if !isDuplicate {
						result.Relationships = append(result.Relationships, rel)
						if !visited[rel.SourceID] {
							nextLevel = append(nextLevel, rel.SourceID)
						}
					}
				}
			}
		}

		currentLevel = nextLevel
	}

	// If no entities were found, return error
	if len(result.Entities) == 0 {
		return nil, ErrEntityNotFound
	}

	s.logger.Debug(ctx, "Graph traversal completed", map[string]interface{}{
		"startEntityId": entityID,
		"depth":         depth,
		"entities":      len(result.Entities),
		"relationships": len(result.Relationships),
	})

	return result, nil
}

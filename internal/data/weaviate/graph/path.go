package graph

import (
	"context"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/multitenancy"
)

// ShortestPath finds the shortest path between two entities using BFS.
func (s *Store) ShortestPath(ctx context.Context, sourceID, targetID string, opts ...contracts.GraphSearchOption) (*contracts.GraphPath, error) {
	if sourceID == "" || targetID == "" {
		return nil, ErrInvalidEntityID
	}

	if sourceID == targetID {
		// Same entity - return empty path
		entity, err := s.GetEntity(ctx, sourceID)
		if err != nil {
			return nil, err
		}
		return &contracts.GraphPath{
			Source:        *entity,
			Target:        *entity,
			Entities:      []contracts.Entity{},
			Relationships: []contracts.Relationship{},
			Length:        0,
		}, nil
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

	// BFS to find shortest path
	maxDepth := options.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 5 // Default max depth for path finding
	}

	type pathNode struct {
		entityID string
		path     []string
		rels     []string
	}

	visited := make(map[string]bool)
	queue := []pathNode{{entityID: sourceID, path: []string{sourceID}, rels: []string{}}}

	searchOpts := []contracts.GraphSearchOption{}
	if tenant != "" {
		searchOpts = append(searchOpts, contracts.WithSearchTenant(tenant))
	}
	if len(options.RelationshipTypes) > 0 {
		searchOpts = append(searchOpts, contracts.WithRelationshipTypes(options.RelationshipTypes...))
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.entityID == targetID {
			// Found the path - build the result
			return s.buildPathResult(ctx, current.path, current.rels, tenant)
		}

		if visited[current.entityID] {
			continue
		}
		visited[current.entityID] = true

		// Check depth limit
		if len(current.path) > maxDepth {
			continue
		}

		// Get relationships
		rels, err := s.GetRelationships(ctx, current.entityID, contracts.DirectionBoth, searchOpts...)
		if err != nil {
			continue
		}

		for _, rel := range rels {
			nextID := rel.TargetID
			if rel.TargetID == current.entityID {
				nextID = rel.SourceID
			}

			if !visited[nextID] {
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = nextID

				newRels := make([]string, len(current.rels)+1)
				copy(newRels, current.rels)
				newRels[len(current.rels)] = rel.ID

				queue = append(queue, pathNode{
					entityID: nextID,
					path:     newPath,
					rels:     newRels,
				})
			}
		}
	}

	return nil, ErrPathNotFound
}

// buildPathResult constructs a GraphPath from entity and relationship IDs.
func (s *Store) buildPathResult(ctx context.Context, entityIDs []string, relIDs []string, tenant string) (*contracts.GraphPath, error) {
	if len(entityIDs) < 2 {
		return nil, ErrPathNotFound
	}

	storeOpts := []contracts.GraphStoreOption{}
	if tenant != "" {
		storeOpts = append(storeOpts, contracts.WithGraphTenant(tenant))
	}

	// Get source entity
	source, err := s.GetEntity(ctx, entityIDs[0], storeOpts...)
	if err != nil {
		return nil, err
	}

	// Get target entity
	target, err := s.GetEntity(ctx, entityIDs[len(entityIDs)-1], storeOpts...)
	if err != nil {
		return nil, err
	}

	// Get intermediate entities
	intermediate := []contracts.Entity{}
	for i := 1; i < len(entityIDs)-1; i++ {
		entity, err := s.GetEntity(ctx, entityIDs[i], storeOpts...)
		if err != nil {
			s.logger.Warn(ctx, "Failed to get intermediate entity", map[string]interface{}{
				"entityId": entityIDs[i],
				"error":    err.Error(),
			})
			continue
		}
		intermediate = append(intermediate, *entity)
	}

	// Get relationships
	relationships := []contracts.Relationship{}
	for _, relID := range relIDs {
		// We need to find the relationship by ID
		// This is a limitation - we'll try to find it from the stored relationships
		// For now, we'll create a placeholder
		rel := contracts.Relationship{ID: relID}
		relationships = append(relationships, rel)
	}

	return &contracts.GraphPath{
		Source:        *source,
		Target:        *target,
		Entities:      intermediate,
		Relationships: relationships,
		Length:        len(entityIDs) - 1,
	}, nil
}

package vector

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"

	"nu/internal/contracts"
)

// Search searches for similar documents
func (s *Store) Search(ctx context.Context, query string, limit int, options ...contracts.SearchOption) ([]contracts.SearchResult, error) {
	// Apply options
	opts := &contracts.SearchOptions{
		MinScore: 0.0,
	}
	for _, option := range options {
		option(opts)
	}

	// Get class name
	className, err := s.getClassName(ctx, opts.Class)
	if err != nil {
		return nil, err
	}

	// Generate embedding for the query
	vector, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for query: %w", err)
	}

	// Build query
	whereFilter := s.buildWhereFilter(opts.Filters)

	// Debug log for filter
	if len(opts.Filters) > 0 {
		s.logger.Info(ctx, "Applying filters", map[string]interface{}{"filters": opts.Filters})
		if whereFilter != nil {
			s.logger.Info(ctx, "Built where filter", map[string]interface{}{"filter": whereFilter})
		} else {
			s.logger.Info(ctx, "Warning: Failed to build where filter from filters", nil)
		}
	}

	// Log the GraphQL query details
	s.logger.Info(ctx, "Executing GraphQL query", map[string]interface{}{
		"className": className,
		"limit":     limit,
		"query":     query,
	})

	// Build dynamic field list
	fieldList, err := s.buildFieldList(ctx, className, opts.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to build field list: %w", err)
	}

	s.logger.Debug(ctx, "Using field list for search", map[string]interface{}{
		"fieldList": fieldList,
		"className": className,
	})

	// Build query with dynamic fields
	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(graphql.Field{
			Name: fieldList,
		}).
		WithNearVector(s.client.GraphQL().NearVectorArgBuilder().
			WithVector(vector)).
		WithLimit(limit)

	// Add where filter if specified
	if whereFilter != nil {
		queryBuilder = queryBuilder.WithWhere(whereFilter)
	}

	// Add tenant support if specified
	if opts.Tenant != "" {
		queryBuilder = queryBuilder.WithTenant(opts.Tenant)
	}

	result, err := queryBuilder.Do(ctx)

	if err != nil {
		s.logger.Error(ctx, "GraphQL query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	// Log the raw response for debugging
	s.logger.Info(ctx, "GraphQL response received", map[string]interface{}{
		"rawData": result.Data,
		"errors":  result.Errors,
	})

	// Parse results
	searchResults, err := s.parseSearchResults(result, className)
	if err != nil {
		return nil, err
	}

	// Apply similarity threshold
	filteredResults := []contracts.SearchResult{}
	for _, res := range searchResults {
		if res.Score >= opts.MinScore {
			filteredResults = append(filteredResults, res)
		}
	}

	return filteredResults, nil
}

// SearchByVector searches for similar documents using a vector
func (s *Store) SearchByVector(ctx context.Context, vector []float32, limit int, options ...contracts.SearchOption) ([]contracts.SearchResult, error) {
	// Apply options
	opts := &contracts.SearchOptions{
		MinScore: 0.0,
	}
	for _, option := range options {
		option(opts)
	}

	// Get class name
	className, err := s.getClassName(ctx, opts.Class)
	if err != nil {
		return nil, err
	}

	// Build query
	whereFilter := s.buildWhereFilter(opts.Filters)

	// Build dynamic field list
	fieldList, err := s.buildFieldList(ctx, className, opts.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to build field list: %w", err)
	}

	s.logger.Debug(ctx, "Using field list for vector search", map[string]interface{}{
		"fieldList": fieldList,
		"className": className,
	})

	// Use vector search
	queryBuilder := s.client.GraphQL().Get().
		WithClassName(className).
		WithFields(graphql.Field{
			Name: fieldList,
		}).
		WithNearVector(s.client.GraphQL().NearVectorArgBuilder().
			WithVector(vector)).
		WithWhere(whereFilter).
		WithLimit(limit)

	// Add tenant support if specified
	if opts.Tenant != "" {
		queryBuilder = queryBuilder.WithTenant(opts.Tenant)
	}

	result, err := queryBuilder.Do(ctx)
	if err != nil {
		s.logger.Error(ctx, "GraphQL query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute vector search: %w", err)
	}

	// Parse results
	return s.parseSearchResults(result, className)
}

// GlobalSearch searches for documents without tenant context (for shared data)
func (s *Store) GlobalSearch(ctx context.Context, query string, limit int, options ...contracts.SearchOption) ([]contracts.SearchResult, error) {
	// Create a context without organization ID to ensure global search
	globalCtx := context.Background()
	return s.Search(globalCtx, query, limit, options...)
}

// GlobalSearchByVector searches for documents by vector without tenant context (for shared data)
func (s *Store) GlobalSearchByVector(ctx context.Context, vector []float32, limit int, options ...contracts.SearchOption) ([]contracts.SearchResult, error) {
	// Create a context without organization ID to ensure global search
	globalCtx := context.Background()
	return s.SearchByVector(globalCtx, vector, limit, options...)
}

// buildFieldList constructs the GraphQL field specification for queries
// If fields are specified in options, uses those; otherwise discovers all fields from schema
func (s *Store) buildFieldList(ctx context.Context, className string, fields []string) (string, error) {
	// If specific fields are requested, use them
	if len(fields) > 0 {
		fieldList := ""
		for i, field := range fields {
			if i > 0 {
				fieldList += " "
			}
			fieldList += field
		}
		// Always include _additional metadata
		fieldList += " _additional { certainty id }"
		return fieldList, nil
	}

	// Auto-discover all fields from schema
	schema, err := s.client.Schema().Getter().Do(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get schema for field discovery", map[string]interface{}{
			"error":     err.Error(),
			"className": className,
		})
		// Fallback to basic fields if schema discovery fails
		return "content _additional { certainty id }", nil
	}

	// Find the target class
	var targetClass *models.Class
	for _, class := range schema.Classes {
		if class.Class == className {
			targetClass = class
			break
		}
	}

	if targetClass == nil {
		s.logger.Warn(ctx, "Class not found in schema, using fallback fields", map[string]interface{}{
			"className": className,
		})
		// Fallback to basic fields if class not found
		return "content _additional { certainty id }", nil
	}

	// Build field list from all properties
	fieldList := ""
	for i, property := range targetClass.Properties {
		if i > 0 {
			fieldList += " "
		}
		fieldList += property.Name
	}

	// Always include _additional metadata
	fieldList += " _additional { certainty id }"

	s.logger.Debug(ctx, "Built dynamic field list", map[string]interface{}{
		"className":  className,
		"fieldList":  fieldList,
		"fieldCount": len(targetClass.Properties),
	})

	return fieldList, nil
}

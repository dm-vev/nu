package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/supabase-community/postgrest-go"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// Query queries documents in the collection
func (c *SupabaseCollection) Query(ctx context.Context, filter map[string]interface{}, options ...contracts.QueryOption) ([]map[string]interface{}, error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Apply options
	opts := &contracts.QueryOptions{}
	for _, option := range options {
		option(opts)
	}

	// Start query
	query := c.client.supabase.From(c.name).Select("*", "", false)

	// Add organization ID filter
	query = query.Eq("org_id", orgID)

	// Add filters
	for k, v := range filter {
		query = query.Eq(k, v.(string))
	}

	// Add limit and offset
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit, "")
	}
	// Offset is not supported by this API

	// Add order by
	if opts.OrderBy != "" {
		if strings.ToLower(opts.OrderDirection) == "desc" {
			query = query.Order(opts.OrderBy, &postgrest.OrderOpts{Ascending: false})
		} else {
			query = query.Order(opts.OrderBy, &postgrest.OrderOpts{Ascending: true})
		}
	}

	// Execute query
	resp, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}

	// Parse response
	var results []map[string]interface{}
	if err := json.Unmarshal(resp, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return results, nil
}

package orchestration

import (
	"context"
	"fmt"
	"strings"
)

// SimpleRouter routes requests based on a simple keyword matching
type HandoffSimpleRouter struct {
	routes map[string][]string // maps keywords to agent IDs
}

// NewSimpleRouter creates a new simple router
func NewHandoffSimpleRouter() *HandoffSimpleRouter {
	return &HandoffSimpleRouter{
		routes: make(map[string][]string),
	}
}

// AddRoute adds a route to the router
func (r *HandoffSimpleRouter) AddRoute(keyword string, agentID string) {
	r.routes[keyword] = append(r.routes[keyword], agentID)
}

// Route determines which agent should handle a request
func (r *HandoffSimpleRouter) Route(ctx context.Context, query string, context map[string]interface{}) (string, error) {
	// Simple keyword matching
	for keyword, agentIDs := range r.routes {
		if contains(query, keyword) {
			// Return the first agent ID
			if len(agentIDs) > 0 {
				return agentIDs[0], nil
			}
		}
	}

	return "", fmt.Errorf("no agent found for query: %s", query)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

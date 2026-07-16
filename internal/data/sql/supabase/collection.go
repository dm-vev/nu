package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dm-vev/nu/internal/multitenancy"
)

// Collection represents a reference to a collection/table
type Collection struct {
	client *Client
	name   string
}

// Insert inserts a document into the collection
func (c *Collection) Insert(ctx context.Context, data map[string]interface{}) (string, error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Add organization ID and created_at to data
	data["org_id"] = orgID
	data["created_at"] = time.Now()

	// Generate ID if not provided
	id, ok := data["id"].(string)
	if !ok || id == "" {
		id = uuid.New().String()
		data["id"] = id
	}

	// Insert data
	resp, _, err := c.client.supabase.From(c.name).Insert(data, false, "", "", "").Execute()
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	// Parse response to check for success
	var results []map[string]interface{}
	if err := json.Unmarshal(resp, &results); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no document was inserted")
	}

	return id, nil
}

// Get retrieves a document by ID
func (c *Collection) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Query document
	resp, _, err := c.client.supabase.From(c.name).
		Select("*", "", false).
		Eq("id", id).
		Eq("org_id", orgID).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Parse response
	var results []map[string]interface{}
	if err := json.Unmarshal(resp, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("document not found")
	}

	return results[0], nil
}

// Update updates a document by ID
func (c *Collection) Update(ctx context.Context, id string, data map[string]interface{}) error {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Add updated_at to data
	data["updated_at"] = time.Now()

	// Update document
	resp, _, err := c.client.supabase.From(c.name).
		Update(data, "", "").
		Eq("id", id).
		Eq("org_id", orgID).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	// Parse response to check for success
	var results []map[string]interface{}
	if err := json.Unmarshal(resp, &results); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no document was updated")
	}

	return nil
}

// Delete deletes a document by ID
func (c *Collection) Delete(ctx context.Context, id string) error {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Delete document
	resp, _, err := c.client.supabase.From(c.name).
		Delete("", "").
		Eq("id", id).
		Eq("org_id", orgID).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Parse response to check for success
	var results []map[string]interface{}
	if err := json.Unmarshal(resp, &results); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no document was deleted")
	}

	return nil
}

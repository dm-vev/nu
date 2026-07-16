package vector

import (
	"context"
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/weaviate/weaviate/entities/models"

	"nu/internal/contracts"
)

// Store stores documents in Weaviate with optional tenant support
func (s *Store) Store(ctx context.Context, documents []contracts.Document, options ...contracts.StoreOption) error {
	// Apply options
	opts := &contracts.StoreOptions{
		BatchSize: 100,
	}
	for _, option := range options {
		option(opts)
	}

	// Get class name
	className, err := s.getClassName(ctx, opts.Class)
	if err != nil {
		return err
	}

	// Store documents in batches
	batch := s.client.Batch().ObjectsBatcher()
	batchSize := opts.BatchSize
	batchCount := 0

	for _, doc := range documents {
		// Generate embedding for the document content
		vector, err := s.embedder.Embed(ctx, doc.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		properties := map[string]interface{}{
			"content": doc.Content,
		}
		for k, v := range doc.Metadata {
			properties[k] = v
		}

		obj := &models.Object{
			Class:      className,
			ID:         strfmt.UUID(doc.ID),
			Properties: properties,
			Vector:     vector, // Use the generated vector
		}

		// Add tenant support if specified
		if opts.Tenant != "" {
			obj.Tenant = opts.Tenant
		}

		batch.WithObjects(obj)
		batchCount++

		// Execute batch when it reaches the batch size
		if batchCount >= batchSize {
			if _, err := batch.Do(ctx); err != nil {
				return fmt.Errorf("failed to store batch: %w", err)
			}
			// Reset batch and count
			batch = s.client.Batch().ObjectsBatcher()
			batchCount = 0
		}
	}

	// Final batch
	if batchCount > 0 {
		if _, err := batch.Do(ctx); err != nil {
			return fmt.Errorf("failed to store final batch: %w", err)
		}
	}

	return nil
}

// Delete removes documents from Weaviate
func (s *Store) Delete(ctx context.Context, ids []string, options ...contracts.DeleteOption) error {
	// Apply options
	opts := &contracts.DeleteOptions{}
	for _, option := range options {
		option(opts)
	}

	// Get class name
	className, err := s.getClassName(ctx, opts.Class)
	if err != nil {
		return err
	}

	// Delete objects
	for _, id := range ids {
		deleter := s.client.Data().Deleter().
			WithClassName(className).
			WithID(id)

		// Add tenant support if specified
		if opts.Tenant != "" {
			deleter = deleter.WithTenant(opts.Tenant)
		}

		if err := deleter.Do(ctx); err != nil {
			return fmt.Errorf("failed to delete document %s: %w", id, err)
		}
	}

	return nil
}

// Get retrieves a single document by ID
func (s *Store) Get(ctx context.Context, id string, options ...contracts.StoreOption) (*contracts.Document, error) {
	// Apply options
	opts := &contracts.StoreOptions{}
	for _, option := range options {
		option(opts)
	}

	// Get class name (use default since we're getting by ID)
	className, err := s.getClassName(ctx, opts.Class)
	if err != nil {
		return nil, err
	}

	getter := s.client.Data().ObjectsGetter().
		WithClassName(className).
		WithID(id)

	// Add tenant support if specified
	if opts.Tenant != "" {
		getter = getter.WithTenant(opts.Tenant)
	}

	result, err := getter.Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get document %s: %w", id, err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("document %s not found", id)
	}

	doc := &contracts.Document{
		ID:       id,
		Content:  result[0].Properties.(map[string]interface{})["content"].(string),
		Metadata: make(map[string]interface{}),
	}

	// Copy all properties except content to metadata
	for k, v := range result[0].Properties.(map[string]interface{}) {
		if k != "content" {
			doc.Metadata[k] = v
		}
	}

	return doc, nil
}

// GlobalStore stores documents in Weaviate without tenant context (for shared data)
func (s *Store) GlobalStore(ctx context.Context, documents []contracts.Document, options ...contracts.StoreOption) error {
	// Create a context without organization ID to ensure global storage
	globalCtx := context.Background()
	return s.Store(globalCtx, documents, options...)
}

// GlobalDelete deletes documents without tenant context (for shared data)
func (s *Store) GlobalDelete(ctx context.Context, ids []string, options ...contracts.DeleteOption) error {
	// Create a context without organization ID to ensure global deletion
	globalCtx := context.Background()
	return s.Delete(globalCtx, ids, options...)
}

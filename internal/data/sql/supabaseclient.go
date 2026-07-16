package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/supabase-community/supabase-go"

	"nu/internal/contracts"
)

// Client implements the DataStore interface for Supabase
type SupabaseClient struct {
	supabase *supabase.Client
	db       *sql.DB
}

// Option represents an option for configuring the client
type SupabaseOption func(*SupabaseClient)

// WithDB sets the SQL database connection
func WithSupabaseDB(db *sql.DB) SupabaseOption {
	return func(c *SupabaseClient) {
		c.db = db
	}
}

// New creates a new Supabase client
func NewSupabase(url string, apiKey string, options ...SupabaseOption) (*SupabaseClient, error) {
	supabaseClient, err := supabase.NewClient(url, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	client := &SupabaseClient{
		supabase: supabaseClient,
	}

	for _, option := range options {
		option(client)
	}

	return client, nil
}

// Collection returns a reference to a specific collection/table
func (c *SupabaseClient) Collection(name string) contracts.CollectionRef {
	return &SupabaseCollection{
		client: c,
		name:   name,
	}
}

// Transaction executes multiple operations in a transaction
func (c *SupabaseClient) Transaction(ctx context.Context, fn func(tx contracts.Transaction) error) error {
	if c.db == nil {
		return errors.New("database connection is required for transactions")
	}

	// Start transaction
	sqlTx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Create transaction object
	tx := &SupabaseTransaction{
		client: c,
		tx:     sqlTx,
	}

	// Execute transaction function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed with error: %v, rollback failed with error: %w", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection
func (c *SupabaseClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

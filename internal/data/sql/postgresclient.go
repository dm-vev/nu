package sql

import (
	"database/sql"
	"fmt"

	"nu/internal/contracts"
)

// Client implements the DataStore interface for PostgreSQL
type PostgresClient struct {
	db *sql.DB
}

// Option represents an option for configuring the client
type PostgresOption func(*PostgresClient)

// New creates a new PostgreSQL client
func NewPostgres(connectionString string, options ...PostgresOption) (*PostgresClient, error) {
	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	client := &PostgresClient{
		db: db,
	}

	for _, option := range options {
		option(client)
	}

	return client, nil
}

// NewWithDB creates a new PostgreSQL client with an existing database connection
func NewPostgresWithDB(db *sql.DB) (*PostgresClient, error) {
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresClient{
		db: db,
	}, nil
}

// Collection returns a reference to a specific collection/table
func (c *PostgresClient) Collection(name string) contracts.CollectionRef {
	return &PostgresCollection{
		client: c,
		name:   name,
	}
}

// Close closes the database connection
func (c *PostgresClient) Close() error {
	return c.db.Close()
}

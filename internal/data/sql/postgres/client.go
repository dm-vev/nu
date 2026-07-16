package postgres

import (
	"database/sql"
	"fmt"

	"github.com/dm-vev/nu/contracts"
)

// Client implements the DataStore interface for PostgreSQL
type Client struct {
	db *sql.DB
}

// Option represents an option for configuring the client
type Option func(*Client)

// New creates a new PostgreSQL client
func New(connectionString string, options ...Option) (*Client, error) {
	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	client := &Client{
		db: db,
	}

	for _, option := range options {
		option(client)
	}

	return client, nil
}

// NewWithDB creates a new PostgreSQL client with an existing database connection
func NewWithDB(db *sql.DB) (*Client, error) {
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &Client{
		db: db,
	}, nil
}

// Collection returns a reference to a specific collection/table
func (c *Client) Collection(name string) contracts.CollectionRef {
	return &Collection{
		client: c,
		name:   name,
	}
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

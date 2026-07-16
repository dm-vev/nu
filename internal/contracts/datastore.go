package contracts

import (
	"context"
)

// DataStore represents a database for storing structured data
type DataStore interface {
	// Collection returns a reference to a specific collection/table
	Collection(name string) CollectionRef

	// Transaction executes multiple operations in a transaction
	Transaction(ctx context.Context, fn func(tx Transaction) error) error

	// Close closes the database connection
	Close() error
}

// CollectionRef represents a reference to a collection/table
type CollectionRef interface {
	// Insert inserts a document into the collection
	Insert(ctx context.Context, data map[string]interface{}) (string, error)

	// Get retrieves a document by ID
	Get(ctx context.Context, id string) (map[string]interface{}, error)

	// Update updates a document by ID
	Update(ctx context.Context, id string, data map[string]interface{}) error

	// Delete deletes a document by ID
	Delete(ctx context.Context, id string) error

	// Query queries documents in the collection
	Query(ctx context.Context, filter map[string]interface{}, options ...QueryOption) ([]map[string]interface{}, error)
}

// Transaction represents a database transaction
type Transaction interface {
	// Collection returns a reference to a specific collection/table within the transaction
	Collection(name string) CollectionRef

	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error
}

// QueryOptions contains options for querying documents
type QueryOptions struct {
	// Limit is the maximum number of documents to return
	Limit int

	// Offset is the number of documents to skip
	Offset int

	// OrderBy specifies the field to order by
	OrderBy string

	// OrderDirection specifies the order direction (asc or desc)
	OrderDirection string
}

// QueryOption represents an option for querying documents
type QueryOption func(*QueryOptions)

// QueryWithLimit sets the maximum number of documents to return
func QueryWithLimit(limit int) QueryOption {
	return func(o *QueryOptions) {
		o.Limit = limit
	}
}

// QueryWithOffset sets the number of documents to skip
func QueryWithOffset(offset int) QueryOption {
	return func(o *QueryOptions) {
		o.Offset = offset
	}
}

// QueryWithOrderBy sets the field to order by and the direction
func QueryWithOrderBy(field string, direction string) QueryOption {
	return func(o *QueryOptions) {
		o.OrderBy = field
		o.OrderDirection = direction
	}
}

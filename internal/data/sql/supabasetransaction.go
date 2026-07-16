package sql

import (
	"database/sql"

	"nu/internal/contracts"
)

// Transaction represents a database transaction
type SupabaseTransaction struct {
	client *SupabaseClient
	tx     *sql.Tx
}

// Collection returns a reference to a specific collection/table within the transaction
func (t *SupabaseTransaction) Collection(name string) contracts.CollectionRef {
	return &SupabaseTransactionCollection{
		tx:   t.tx,
		name: name,
	}
}

// Commit commits the transaction
func (t *SupabaseTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *SupabaseTransaction) Rollback() error {
	return t.tx.Rollback()
}

package supabase

import (
	"database/sql"

	"github.com/dm-vev/nu/contracts"
)

// Transaction represents a database transaction
type Transaction struct {
	client *Client
	tx     *sql.Tx
}

// Collection returns a reference to a specific collection/table within the transaction
func (t *Transaction) Collection(name string) contracts.CollectionRef {
	return &TransactionCollection{
		tx:   t.tx,
		name: name,
	}
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

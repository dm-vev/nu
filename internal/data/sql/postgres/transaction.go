package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dm-vev/nu/contracts"
)

// Transaction executes multiple operations in a transaction
func (c *Client) Transaction(ctx context.Context, fn func(tx contracts.Transaction) error) error {
	// Start transaction
	sqlTx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Create transaction object
	tx := &Transaction{
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

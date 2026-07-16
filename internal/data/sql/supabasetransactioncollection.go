package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// TransactionCollection represents a collection within a transaction
type SupabaseTransactionCollection struct {
	tx   *sql.Tx
	name string
}

// Insert inserts a document into the collection within a transaction
func (c *SupabaseTransactionCollection) Insert(ctx context.Context, data map[string]interface{}) (string, error) {
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

	// Build query
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	i := 1

	for k, v := range data {
		columns = append(columns, pq.QuoteIdentifier(k))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, v)
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id", // #nosec G201
		pq.QuoteIdentifier(c.name),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Execute query
	var returnedID string
	err = c.tx.QueryRowContext(ctx, query, values...).Scan(&returnedID)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	return returnedID, nil
}

// Get retrieves a document by ID within a transaction
func (c *SupabaseTransactionCollection) Get(ctx context.Context, id string) (map[string]interface{}, error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Build query
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE id = $1 AND org_id = $2",
		pq.QuoteIdentifier(c.name),
	)

	// Execute query
	var result map[string]interface{}
	err = c.tx.QueryRowContext(ctx, query, id, orgID).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return result, nil
}

// Update updates a document by ID within a transaction
func (c *SupabaseTransactionCollection) Update(ctx context.Context, id string, data map[string]interface{}) error {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Add updated_at to data
	data["updated_at"] = time.Now()

	// Build query
	setStatements := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+2)
	i := 1

	for k, v := range data {
		setStatements = append(setStatements, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(k), i))
		values = append(values, v)
		i++
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d AND org_id = $%d", // #nosec G201
		pq.QuoteIdentifier(c.name),
		strings.Join(setStatements, ", "),
		i,
		i+1,
	)

	values = append(values, id, orgID)

	// Execute query
	result, err := c.tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found or not owned by organization")
	}

	return nil
}

// Delete deletes a document by ID within a transaction
func (c *SupabaseTransactionCollection) Delete(ctx context.Context, id string) error {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Build query
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE id = $1 AND org_id = $2",
		pq.QuoteIdentifier(c.name),
	)

	// Execute query
	result, err := c.tx.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found or not owned by organization")
	}

	return nil
}

// Query queries documents in the collection within a transaction
func (c *SupabaseTransactionCollection) Query(ctx context.Context, filter map[string]interface{}, options ...contracts.QueryOption) ([]map[string]interface{}, error) {
	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization ID: %w", err)
	}

	// Apply options
	opts := &contracts.QueryOptions{}
	for _, option := range options {
		option(opts)
	}

	// Build query
	whereStatements := []string{"org_id = $1"}
	values := []interface{}{orgID}
	i := 2

	for k, v := range filter {
		whereStatements = append(whereStatements, fmt.Sprintf("%s = $%d", pq.QuoteIdentifier(k), i))
		values = append(values, v)
		i++
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s", // #nosec G201 - table name is sanitized with pq.QuoteIdentifier and WHERE conditions use parameterized queries
		pq.QuoteIdentifier(c.name),
		strings.Join(whereStatements, " AND "),
	)

	// Add order by
	if opts.OrderBy != "" {
		direction := "ASC"
		if strings.ToLower(opts.OrderDirection) == "desc" {
			direction = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, direction)
	}

	// Add limit and offset
	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}
	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}

	// Execute query
	rows, err := c.tx.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			// Merge with existing error or set if none
			if err == nil {
				err = fmt.Errorf("failed to close rows: %w", cerr)
			}
		}
	}()

	// Parse results
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			result[col] = values[i]
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

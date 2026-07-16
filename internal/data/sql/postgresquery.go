package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// Query queries documents in the collection
func (c *PostgresCollection) Query(ctx context.Context, filter map[string]interface{}, options ...contracts.QueryOption) ([]map[string]interface{}, error) {
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
		query += fmt.Sprintf(" ORDER BY %s %s", pq.QuoteIdentifier(opts.OrderBy), direction)
	}

	// Add limit and offset
	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}
	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}

	// Execute query
	rows, err := c.client.db.QueryContext(ctx, query, values...)
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

// Query queries documents in the collection within a transaction
func (c *PostgresTransactionCollection) Query(ctx context.Context, filter map[string]interface{}, options ...contracts.QueryOption) ([]map[string]interface{}, error) {
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
		query += fmt.Sprintf(" ORDER BY %s %s", pq.QuoteIdentifier(opts.OrderBy), direction)
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

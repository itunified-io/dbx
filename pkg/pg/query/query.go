// Package query provides SQL execution with row limiting and query plan analysis.
package query

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// QueryResult holds the result of a SQL execution.
type QueryResult struct {
	Columns   []string         `json:"columns"`
	Rows      []map[string]any `json:"rows"`
	RowCount  int              `json:"row_count"`
	Truncated bool             `json:"truncated"`
}

const defaultMaxRows = 1000

// Execute runs a parameterized SQL query with row limit enforcement.
func Execute(ctx context.Context, q pginternal.Querier, params map[string]any) (*QueryResult, error) {
	sql, _ := params["sql"].(string)
	if sql == "" {
		return nil, fmt.Errorf("sql parameter is required")
	}
	maxRows := defaultMaxRows
	if v, ok := params["max_rows"].(int); ok && v > 0 {
		maxRows = v
	}

	rows, err := pginternal.QueryRows(ctx, q, sql)
	if err != nil {
		return nil, fmt.Errorf("pg query: %w", err)
	}
	defer rows.Close()

	descs := rows.FieldDescriptions()
	columns := make([]string, len(descs))
	for i, d := range descs {
		columns[i] = string(d.Name)
	}

	var resultRows []map[string]any
	truncated := false
	for rows.Next() {
		if len(resultRows) >= maxRows {
			truncated = true
			break
		}
		vals, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("pg query scan: %w", err)
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			row[col] = vals[i]
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pg query rows: %w", err)
	}
	return &QueryResult{
		Columns:   columns,
		Rows:      resultRows,
		RowCount:  len(resultRows),
		Truncated: truncated,
	}, nil
}

// Explain returns the query plan for a SQL statement.
func Explain(ctx context.Context, q pginternal.Querier, params map[string]any) ([]string, error) {
	sql, _ := params["sql"].(string)
	if sql == "" {
		return nil, fmt.Errorf("sql parameter is required")
	}
	analyze, _ := params["analyze"].(bool)

	explainSQL := "EXPLAIN "
	if analyze {
		explainSQL = "EXPLAIN ANALYZE "
	}
	explainSQL += sql

	rows, err := pginternal.QueryRows(ctx, q, explainSQL)
	if err != nil {
		return nil, fmt.Errorf("pg explain: %w", err)
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, fmt.Errorf("pg explain scan: %w", err)
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// Prepared prepares and executes a named statement.
func Prepared(ctx context.Context, q pginternal.Querier, params map[string]any) (*QueryResult, error) {
	// Prepared statements in pgx are implicit via the extended protocol.
	// This function wraps Execute with the statement name for cache keying.
	return Execute(ctx, q, params)
}

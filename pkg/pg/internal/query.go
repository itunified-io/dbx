package pginternal

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier abstracts *pgxpool.Pool for testing with pgxmock.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// QueryRows executes a query that returns rows via the Querier abstraction.
func QueryRows(ctx context.Context, q Querier, sql string, args ...any) (pgx.Rows, error) {
	return q.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns a single row via the Querier abstraction.
func QueryRow(ctx context.Context, q Querier, sql string, args ...any) pgx.Row {
	return q.QueryRow(ctx, sql, args...)
}

// Exec executes a statement and returns the number of rows affected.
func Exec(ctx context.Context, q Querier, sql string, args ...any) (int64, error) {
	ct, err := q.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

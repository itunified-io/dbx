// Package sql provides read-only SQL execution with a SELECT-only guard.
package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// blockedPrefixes are SQL statement types rejected by the read-only guard.
var blockedPrefixes = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE",
	"TRUNCATE", "MERGE", "GRANT", "REVOKE", "PURGE",
	"BEGIN", "DECLARE", "EXEC", "CALL",
}

// ReadOnlyGuard ensures the statement is a SELECT (not SELECT FOR UPDATE).
func ReadOnlyGuard(stmt string) error {
	upper := strings.TrimSpace(strings.ToUpper(stmt))

	for _, prefix := range blockedPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return fmt.Errorf("blocked: %s statements are not allowed in read-only mode", prefix)
		}
	}

	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "WITH") && !strings.HasPrefix(upper, "EXPLAIN") {
		return fmt.Errorf("blocked: only SELECT/WITH/EXPLAIN statements are allowed")
	}

	if strings.Contains(upper, "FOR UPDATE") {
		return fmt.Errorf("blocked: SELECT FOR UPDATE is not allowed in read-only mode")
	}

	return nil
}

// Exec runs a read-only SELECT statement after the guard check.
func Exec(ctx context.Context, db *sql.DB, stmt string) ([]map[string]any, error) {
	if err := ReadOnlyGuard(stmt); err != nil {
		return nil, err
	}
	return dbinternal.QueryRows(ctx, db, stmt)
}

// ExplainPlanSQL wraps a statement in EXPLAIN PLAN.
const ExplainPlanSQL = `EXPLAIN PLAN FOR `

// ExplainPlan generates an execution plan for a SELECT statement.
func ExplainPlan(ctx context.Context, db *sql.DB, stmt string) ([]map[string]any, error) {
	if err := ReadOnlyGuard(stmt); err != nil {
		return nil, err
	}
	// Execute EXPLAIN PLAN FOR ...
	_, err := db.ExecContext(ctx, ExplainPlanSQL+stmt)
	if err != nil {
		return nil, fmt.Errorf("explain plan: %w", err)
	}
	// Read the plan table
	return dbinternal.QueryRows(ctx, db,
		`SELECT plan_table_output FROM TABLE(DBMS_XPLAN.DISPLAY('PLAN_TABLE', NULL, 'ALL'))`)
}

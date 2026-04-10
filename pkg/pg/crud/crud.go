// Package crud provides confirm-gated CRUD operations for PostgreSQL.
package crud

import (
	"context"
	"fmt"
	"sort"
	"strings"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// MutationResult holds the outcome of a CRUD operation.
type MutationResult struct {
	Operation    string `json:"operation"`
	RowsAffected int64  `json:"rows_affected"`
	Table        string `json:"table"`
}

// Insert performs a confirm-gated INSERT.
func Insert(ctx context.Context, q pginternal.Querier, params map[string]any) (*MutationResult, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute INSERT")
	}
	schema, table := extractTarget(params)
	data, _ := params["data"].(map[string]any)
	if len(data) == 0 {
		return nil, fmt.Errorf("data parameter is required")
	}

	cols, vals, placeholders := buildInsertParts(data)
	sql := fmt.Sprintf(`INSERT INTO %q.%q (%s) VALUES (%s)`,
		schema, table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	affected, err := pginternal.Exec(ctx, q, sql, vals...)
	if err != nil {
		return nil, fmt.Errorf("pg insert: %w", err)
	}
	return &MutationResult{Operation: "INSERT", RowsAffected: affected, Table: table}, nil
}

// Update performs a confirm-gated UPDATE.
func Update(ctx context.Context, q pginternal.Querier, params map[string]any) (*MutationResult, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute UPDATE")
	}
	schema, table := extractTarget(params)
	data, _ := params["data"].(map[string]any)
	where, _ := params["where"].(string)
	if where == "" {
		return nil, fmt.Errorf("where clause is required for UPDATE")
	}

	setClauses, vals := buildSetClauses(data)
	whereArgs, _ := params["args"].([]any)
	allArgs := append(vals, whereArgs...)

	sql := fmt.Sprintf(`UPDATE %q.%q SET %s WHERE %s`,
		schema, table, strings.Join(setClauses, ", "), where)

	affected, err := pginternal.Exec(ctx, q, sql, allArgs...)
	if err != nil {
		return nil, fmt.Errorf("pg update: %w", err)
	}
	return &MutationResult{Operation: "UPDATE", RowsAffected: affected, Table: table}, nil
}

// Delete performs a confirm-gated DELETE. No-WHERE deletes require double-confirm.
func Delete(ctx context.Context, q pginternal.Querier, params map[string]any) (*MutationResult, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute DELETE")
	}
	schema, table := extractTarget(params)
	where, _ := params["where"].(string)

	if where == "" {
		if dconfirm, _ := params["confirm_destructive"].(bool); !dconfirm {
			return nil, fmt.Errorf("double-confirm required: DELETE without WHERE deletes all rows. Set confirm_destructive=true")
		}
	}

	var sql string
	var args []any
	if where != "" {
		sql = fmt.Sprintf(`DELETE FROM %q.%q WHERE %s`, schema, table, where)
		args, _ = params["args"].([]any)
	} else {
		sql = fmt.Sprintf(`DELETE FROM %q.%q`, schema, table)
	}

	affected, err := pginternal.Exec(ctx, q, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("pg delete: %w", err)
	}
	return &MutationResult{Operation: "DELETE", RowsAffected: affected, Table: table}, nil
}

// Upsert performs a confirm-gated INSERT ON CONFLICT DO UPDATE.
func Upsert(ctx context.Context, q pginternal.Querier, params map[string]any) (*MutationResult, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute UPSERT")
	}
	schema, table := extractTarget(params)
	data, _ := params["data"].(map[string]any)
	conflict, _ := params["conflict"].(string)
	if conflict == "" {
		return nil, fmt.Errorf("conflict column is required for UPSERT")
	}

	cols, vals, placeholders := buildInsertParts(data)
	setClauses, _ := buildSetClauses(data)

	sql := fmt.Sprintf(`INSERT INTO %q.%q (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s`,
		schema, table, strings.Join(cols, ", "), strings.Join(placeholders, ", "),
		conflict, strings.Join(setClauses, ", "))

	affected, err := pginternal.Exec(ctx, q, sql, vals...)
	if err != nil {
		return nil, fmt.Errorf("pg upsert: %w", err)
	}
	return &MutationResult{Operation: "UPSERT", RowsAffected: affected, Table: table}, nil
}

// --- helpers ---

func extractTarget(params map[string]any) (string, string) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		schema = "public"
	}
	table, _ := params["table"].(string)
	return schema, table
}

func buildInsertParts(data map[string]any) ([]string, []any, []string) {
	keys := sortedKeys(data)
	cols := make([]string, len(keys))
	vals := make([]any, len(keys))
	phs := make([]string, len(keys))
	for i, k := range keys {
		cols[i] = k
		vals[i] = data[k]
		phs[i] = fmt.Sprintf("$%d", i+1)
	}
	return cols, vals, phs
}

func buildSetClauses(data map[string]any) ([]string, []any) {
	keys := sortedKeys(data)
	clauses := make([]string, len(keys))
	vals := make([]any, len(keys))
	for i, k := range keys {
		clauses[i] = fmt.Sprintf("%s = $%d", k, i+1)
		vals[i] = data[k]
	}
	return clauses, vals
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

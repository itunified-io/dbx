// Package audit provides PostgreSQL audit trail tools.
package audit

import (
	"context"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// QueryLogEntry represents a query log entry.
type QueryLogEntry struct {
	Username  string    `json:"username"`
	Database  string    `json:"database"`
	Query     string    `json:"query"`
	Duration  float64   `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

// ConnectionEntry represents a connection audit entry.
type ConnectionEntry struct {
	PID         int32      `json:"pid"`
	Username    string     `json:"username"`
	Database    string     `json:"database"`
	ClientAddr  string     `json:"client_addr"`
	BackendStart *time.Time `json:"backend_start"`
	State       string     `json:"state"`
}

// PermChange represents a permission change event.
type PermChange struct {
	Grantee       string `json:"grantee"`
	ObjectType    string `json:"object_type"`
	ObjectName    string `json:"object_name"`
	PrivilegeType string `json:"privilege_type"`
	IsGrantable   string `json:"is_grantable"`
}

const sqlConnectionAudit = `
SELECT pid, usename, datname,
       COALESCE(client_addr::text, 'local') AS client_addr,
       backend_start, state
FROM pg_stat_activity
WHERE backend_type = 'client backend'
ORDER BY backend_start DESC`

const sqlPermissionChanges = `
SELECT grantee, table_schema || '.' || table_name AS object_name,
       'TABLE' AS object_type, privilege_type, is_grantable
FROM information_schema.table_privileges
WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY grantee, table_name`

// QueryLog returns recent query log entries (requires pg_stat_statements).
func QueryLog(ctx context.Context, q pginternal.Querier, params map[string]any) ([]QueryLogEntry, error) {
	limit := 50
	if v, ok := params["limit"].(int); ok {
		limit = v
	}
	rows, err := pginternal.QueryRows(ctx, q,
		`SELECT usename, datname, query, total_exec_time AS duration_ms, now() AS timestamp
		 FROM pg_stat_statements ss
		 JOIN pg_stat_activity sa ON sa.pid = pg_backend_pid()
		 ORDER BY total_exec_time DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("pg query log: %w", err)
	}
	defer rows.Close()
	var results []QueryLogEntry
	for rows.Next() {
		var e QueryLogEntry
		if err := rows.Scan(&e.Username, &e.Database, &e.Query, &e.Duration, &e.Timestamp); err != nil {
			return nil, fmt.Errorf("pg query log scan: %w", err)
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

// ConnectionAudit returns all active connections.
func ConnectionAudit(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]ConnectionEntry, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlConnectionAudit)
	if err != nil {
		return nil, fmt.Errorf("pg connection audit: %w", err)
	}
	defer rows.Close()
	var results []ConnectionEntry
	for rows.Next() {
		var e ConnectionEntry
		if err := rows.Scan(&e.PID, &e.Username, &e.Database, &e.ClientAddr,
			&e.BackendStart, &e.State); err != nil {
			return nil, fmt.Errorf("pg connection audit scan: %w", err)
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

// PermissionChanges returns current permission grants.
func PermissionChanges(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]PermChange, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlPermissionChanges)
	if err != nil {
		return nil, fmt.Errorf("pg permission changes: %w", err)
	}
	defer rows.Close()
	var results []PermChange
	for rows.Next() {
		var p PermChange
		if err := rows.Scan(&p.Grantee, &p.ObjectName, &p.ObjectType,
			&p.PrivilegeType, &p.IsGrantable); err != nil {
			return nil, fmt.Errorf("pg permission changes scan: %w", err)
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

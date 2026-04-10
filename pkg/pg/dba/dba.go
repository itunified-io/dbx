// Package dba provides DBA daily operations for PostgreSQL.
package dba

import (
	"context"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// Session represents a backend process from pg_stat_activity.
type Session struct {
	PID             int32      `json:"pid"`
	Username        string     `json:"username"`
	ApplicationName string     `json:"application_name"`
	ClientAddr      string     `json:"client_addr"`
	State           string     `json:"state"`
	WaitEventType   *string    `json:"wait_event_type"`
	WaitEvent       *string    `json:"wait_event"`
	Query           string     `json:"query"`
	QueryStart      *time.Time `json:"query_start"`
	StateChange     *time.Time `json:"state_change"`
	BackendType     string     `json:"backend_type"`
}

// BloatInfo represents table bloat statistics.
type BloatInfo struct {
	SchemaName string  `json:"schema"`
	TableName  string  `json:"table"`
	LiveTuples int64   `json:"n_live_tup"`
	DeadTuples int64   `json:"n_dead_tup"`
	DeadRatio  float64 `json:"dead_ratio"`
	TotalBytes int64   `json:"total_bytes"`
	BloatBytes int64   `json:"bloat_bytes"`
}

// LockDependency represents a blocked/blocking pair.
type LockDependency struct {
	BlockedPID    int32  `json:"blocked_pid"`
	BlockedUser   string `json:"blocked_user"`
	BlockingPID   int32  `json:"blocking_pid"`
	BlockingUser  string `json:"blocking_user"`
	BlockedQuery  string `json:"blocked_query"`
	BlockingQuery string `json:"blocking_query"`
}

const sqlSessionList = `
SELECT pid, usename, application_name,
       COALESCE(client_addr::text, 'local') AS client_addr,
       COALESCE(state, 'unknown') AS state,
       wait_event_type, wait_event, query, query_start, state_change, backend_type
FROM pg_stat_activity
WHERE pid <> pg_backend_pid() AND backend_type = 'client backend'
ORDER BY state, usename`

const sqlBloatCheck = `
SELECT schemaname, relname,
       n_live_tup, n_dead_tup,
       CASE WHEN n_live_tup + n_dead_tup > 0
            THEN ROUND(n_dead_tup::numeric / (n_live_tup + n_dead_tup) * 100, 1)
            ELSE 0 END AS dead_ratio,
       pg_total_relation_size(quote_ident(schemaname)||'.'||quote_ident(relname)) AS total_bytes,
       pg_total_relation_size(quote_ident(schemaname)||'.'||quote_ident(relname)) *
         CASE WHEN n_live_tup + n_dead_tup > 0
              THEN n_dead_tup::numeric / (n_live_tup + n_dead_tup)
              ELSE 0 END AS bloat_bytes
FROM pg_stat_user_tables
WHERE n_dead_tup > 0
ORDER BY dead_ratio DESC`

const sqlLockTree = `
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_query,
       blocking_activity.query AS blocking_query
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted
ORDER BY blocked_activity.query_start`

// SessionList returns all active backend sessions.
func SessionList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]Session, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlSessionList)
	if err != nil {
		return nil, fmt.Errorf("pg session list: %w", err)
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.PID, &s.Username, &s.ApplicationName, &s.ClientAddr,
			&s.State, &s.WaitEventType, &s.WaitEvent, &s.Query,
			&s.QueryStart, &s.StateChange, &s.BackendType); err != nil {
			return nil, fmt.Errorf("pg session list scan: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// SessionKill terminates a backend process. Confirm-gated.
func SessionKill(ctx context.Context, q pginternal.Querier, params map[string]any) (bool, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return false, fmt.Errorf("confirm gate: set confirm=true to terminate backend")
	}
	pid, _ := params["pid"].(int)
	row := pginternal.QueryRow(ctx, q, "SELECT pg_terminate_backend($1)", pid)
	var terminated bool
	if err := row.Scan(&terminated); err != nil {
		return false, fmt.Errorf("pg session kill: %w", err)
	}
	return terminated, nil
}

// SessionCancel cancels a running query. Confirm-gated.
func SessionCancel(ctx context.Context, q pginternal.Querier, params map[string]any) (bool, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return false, fmt.Errorf("confirm gate: set confirm=true to cancel query")
	}
	pid, _ := params["pid"].(int)
	row := pginternal.QueryRow(ctx, q, "SELECT pg_cancel_backend($1)", pid)
	var cancelled bool
	if err := row.Scan(&cancelled); err != nil {
		return false, fmt.Errorf("pg session cancel: %w", err)
	}
	return cancelled, nil
}

// VacuumAnalyze performs VACUUM with optional FULL and ANALYZE. Confirm-gated.
func VacuumAnalyze(ctx context.Context, q pginternal.Querier, params map[string]any) (string, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return "", fmt.Errorf("confirm gate: set confirm=true to execute VACUUM")
	}
	table, _ := params["table"].(string)
	analyze, _ := params["analyze"].(bool)
	full, _ := params["full"].(bool)

	sql := "VACUUM"
	if full {
		sql += " FULL"
	}
	if analyze {
		sql += " ANALYZE"
	}
	if table != "" {
		sql += " " + table
	}

	_, err := pginternal.Exec(ctx, q, sql)
	if err != nil {
		return "", fmt.Errorf("pg vacuum: %w", err)
	}
	fullStr := ""
	if full {
		fullStr = " FULL"
	}
	analyzeStr := ""
	if analyze {
		analyzeStr = " ANALYZE"
	}
	targetStr := "all tables"
	if table != "" {
		targetStr = table
	}
	return fmt.Sprintf("VACUUM%s%s completed for %s", fullStr, analyzeStr, targetStr), nil
}

// Reindex rebuilds an index. Confirm-gated.
func Reindex(ctx context.Context, q pginternal.Querier, params map[string]any) (string, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return "", fmt.Errorf("confirm gate: set confirm=true to execute REINDEX")
	}
	target, _ := params["target"].(string)
	concurrently, _ := params["concurrently"].(bool)

	sql := "REINDEX"
	if concurrently {
		sql += " (CONCURRENTLY)"
	}
	sql += " INDEX " + target

	_, err := pginternal.Exec(ctx, q, sql)
	if err != nil {
		return "", fmt.Errorf("pg reindex: %w", err)
	}
	return fmt.Sprintf("REINDEX completed for %s", target), nil
}

// BloatCheck returns tables with dead tuple ratio above threshold.
func BloatCheck(ctx context.Context, q pginternal.Querier, params map[string]any) ([]BloatInfo, error) {
	threshold := 10.0
	if v, ok := params["threshold"].(float64); ok {
		threshold = v
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlBloatCheck)
	if err != nil {
		return nil, fmt.Errorf("pg bloat check: %w", err)
	}
	defer rows.Close()
	var results []BloatInfo
	for rows.Next() {
		var b BloatInfo
		if err := rows.Scan(&b.SchemaName, &b.TableName, &b.LiveTuples, &b.DeadTuples,
			&b.DeadRatio, &b.TotalBytes, &b.BloatBytes); err != nil {
			return nil, fmt.Errorf("pg bloat scan: %w", err)
		}
		if b.DeadRatio >= threshold {
			results = append(results, b)
		}
	}
	return results, rows.Err()
}

// LockTree returns blocking/blocked lock pairs.
func LockTree(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]LockDependency, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlLockTree)
	if err != nil {
		return nil, fmt.Errorf("pg lock tree: %w", err)
	}
	defer rows.Close()
	var locks []LockDependency
	for rows.Next() {
		var l LockDependency
		if err := rows.Scan(&l.BlockedPID, &l.BlockedUser, &l.BlockingPID,
			&l.BlockingUser, &l.BlockedQuery, &l.BlockingQuery); err != nil {
			return nil, fmt.Errorf("pg lock tree scan: %w", err)
		}
		locks = append(locks, l)
	}
	return locks, rows.Err()
}

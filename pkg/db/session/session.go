// Package session provides read-only Oracle session operations.
package session

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// ListSQL returns active sessions from v$session.
const ListSQL = `SELECT sid, serial#, username, status, machine, program, sql_id,
       last_call_et, event, wait_class
FROM v$session
WHERE type = 'USER' AND username IS NOT NULL
ORDER BY last_call_et DESC`

// DescribeSQL returns detailed session info.
const DescribeSQL = `SELECT sid, serial#, username, status, machine, program,
       module, action, sql_id, prev_sql_id, logon_time,
       last_call_et, event, wait_class, state,
       blocking_session, blocking_session_status
FROM v$session WHERE sid = :1`

// TopWaitersSQL returns top sessions by wait time.
const TopWaitersSQL = `SELECT sid, serial#, username, event, wait_class,
       seconds_in_wait, state, sql_id, machine
FROM v$session
WHERE type = 'USER' AND username IS NOT NULL AND state = 'WAITING'
ORDER BY seconds_in_wait DESC
FETCH FIRST :1 ROWS ONLY`

// List returns all active user sessions.
func List(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ListSQL)
}

// Describe returns details for a specific session.
func Describe(ctx context.Context, db *sql.DB, sid int) (map[string]any, error) {
	return dbinternal.QueryRow(ctx, db, DescribeSQL, sid)
}

// TopWaiters returns the top N sessions by wait time.
func TopWaiters(ctx context.Context, db *sql.DB, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}
	return dbinternal.QueryRows(ctx, db, TopWaitersSQL, limit)
}

// Package redo provides read-only Oracle redo log operations.
package redo

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// ListSQL returns redo log group info.
const ListSQL = `SELECT l.group#, l.thread#, l.sequence#, l.bytes / 1024 / 1024 AS size_mb,
       l.members, l.archived, l.status, l.first_change#, l.first_time
FROM v$log l ORDER BY l.group#`

// SwitchHistorySQL returns recent log switch history.
const SwitchHistorySQL = `SELECT TO_CHAR(first_time, 'YYYY-MM-DD HH24') AS hour,
       COUNT(*) AS switches
FROM v$log_history
WHERE first_time > SYSDATE - :1
GROUP BY TO_CHAR(first_time, 'YYYY-MM-DD HH24')
ORDER BY hour DESC`

// List returns all redo log groups.
func List(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ListSQL)
}

// SwitchHistory returns log switch counts grouped by hour for the last N days.
func SwitchHistory(ctx context.Context, db *sql.DB, days int) ([]map[string]any, error) {
	if days <= 0 {
		days = 7
	}
	return dbinternal.QueryRows(ctx, db, SwitchHistorySQL, days)
}

// Package undo provides read-only Oracle undo/rollback operations.
package undo

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// ListSQL returns undo tablespace info.
const ListSQL = `SELECT tablespace_name, status,
       ROUND(SUM(bytes) / 1024 / 1024) AS size_mb
FROM dba_undo_extents
GROUP BY tablespace_name, status
ORDER BY tablespace_name, status`

// SegmentInfoSQL returns undo segment details.
const SegmentInfoSQL = `SELECT s.segment_name, s.tablespace_name, s.status,
       ROUND(s.size_kb / 1024) AS size_mb,
       t.xacts AS active_transactions
FROM v$rollstat r
JOIN dba_rollback_segs s ON r.usn = s.segment_id
LEFT JOIN v$transaction t ON r.usn = t.xidusn
ORDER BY s.segment_name`

// List returns undo tablespace usage summary.
func List(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ListSQL)
}

// SegmentInfo returns detailed undo segment information.
func SegmentInfo(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, SegmentInfoSQL)
}

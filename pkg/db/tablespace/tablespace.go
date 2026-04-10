// Package tablespace provides read-only Oracle tablespace operations.
package tablespace

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// ListSQL returns tablespace summary with usage metrics.
const ListSQL = `SELECT t.tablespace_name, t.status, t.contents, t.block_size,
       ROUND(d.bytes / 1024 / 1024) AS size_mb,
       ROUND((d.bytes - NVL(f.free_bytes, 0)) / 1024 / 1024) AS used_mb,
       ROUND(NVL(f.free_bytes, 0) / 1024 / 1024) AS free_mb,
       ROUND((d.bytes - NVL(f.free_bytes, 0)) / d.bytes * 100, 1) AS pct_used
FROM dba_tablespaces t
JOIN (SELECT tablespace_name, SUM(bytes) AS bytes FROM dba_data_files GROUP BY tablespace_name) d
  ON t.tablespace_name = d.tablespace_name
LEFT JOIN (SELECT tablespace_name, SUM(bytes) AS free_bytes FROM dba_free_space GROUP BY tablespace_name) f
  ON t.tablespace_name = f.tablespace_name
ORDER BY pct_used DESC`

// DescribeSQL returns detailed tablespace info including datafiles.
const DescribeSQL = `SELECT file_name, file_id, tablespace_name,
       ROUND(bytes / 1024 / 1024) AS size_mb,
       autoextensible,
       ROUND(maxbytes / 1024 / 1024) AS max_size_mb,
       status
FROM dba_data_files WHERE tablespace_name = :1
ORDER BY file_id`

// UsageSummarySQL returns aggregated usage across all tablespaces.
const UsageSummarySQL = `SELECT COUNT(*) AS tablespace_count,
       ROUND(SUM(d.bytes) / 1024 / 1024 / 1024, 1) AS total_gb,
       ROUND(SUM(d.bytes - NVL(f.free_bytes, 0)) / 1024 / 1024 / 1024, 1) AS used_gb,
       ROUND(SUM(NVL(f.free_bytes, 0)) / 1024 / 1024 / 1024, 1) AS free_gb
FROM dba_tablespaces t
JOIN (SELECT tablespace_name, SUM(bytes) AS bytes FROM dba_data_files GROUP BY tablespace_name) d
  ON t.tablespace_name = d.tablespace_name
LEFT JOIN (SELECT tablespace_name, SUM(bytes) AS free_bytes FROM dba_free_space GROUP BY tablespace_name) f
  ON t.tablespace_name = f.tablespace_name`

// List returns all tablespaces with usage metrics.
func List(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ListSQL)
}

// Describe returns datafiles for a specific tablespace.
func Describe(ctx context.Context, db *sql.DB, name string) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, DescribeSQL, name)
}

// UsageSummary returns aggregated usage across all tablespaces.
func UsageSummary(ctx context.Context, db *sql.DB) (map[string]any, error) {
	return dbinternal.QueryRow(ctx, db, UsageSummarySQL)
}

// Package advisor provides read-only Oracle advisor operations.
package advisor

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// SegmentAdvisorSQL returns segment advisor recommendations.
const SegmentAdvisorSQL = `SELECT tablespace_name, segment_owner, segment_name, segment_type,
       ROUND(allocated_space / 1024 / 1024) AS allocated_mb,
       ROUND(used_space / 1024 / 1024) AS used_mb,
       ROUND(reclaimable_space / 1024 / 1024) AS reclaimable_mb,
       recommendations, c1, c2, c3
FROM TABLE(DBMS_SPACE.ASA_RECOMMENDATIONS())
ORDER BY reclaimable_space DESC`

// SQLTuningListSQL returns SQL tuning advisor task summaries.
const SQLTuningListSQL = `SELECT task_id, task_name, advisor_name, status,
       created, last_modified, execution_start, execution_end,
       how_created, description
FROM dba_advisor_tasks
WHERE advisor_name = 'SQL Tuning Advisor'
ORDER BY created DESC
FETCH FIRST :1 ROWS ONLY`

// SegmentAdvisor returns segment advisor recommendations.
func SegmentAdvisor(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, SegmentAdvisorSQL)
}

// SQLTuningList returns recent SQL tuning advisor tasks.
func SQLTuningList(ctx context.Context, db *sql.DB, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	return dbinternal.QueryRows(ctx, db, SQLTuningListSQL, limit)
}

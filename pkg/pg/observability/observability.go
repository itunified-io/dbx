// Package observability provides PostgreSQL monitoring and observability tools.
package observability

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// WaitEvent represents a wait event summary.
type WaitEvent struct {
	WaitEventType string `json:"wait_event_type"`
	WaitEvent     string `json:"wait_event"`
	Count         int64  `json:"count"`
}

// CheckpointInfo represents checkpoint statistics.
type CheckpointInfo struct {
	CheckpointsTimed int64 `json:"checkpoints_timed"`
	CheckpointsReq   int64 `json:"checkpoints_req"`
	BuffersCheckpoint int64 `json:"buffers_checkpoint"`
	BuffersClean     int64 `json:"buffers_clean"`
	BuffersBackend   int64 `json:"buffers_backend"`
	MaxwrittenClean  int64 `json:"maxwritten_clean"`
}

// TableIO represents table I/O statistics.
type TableIO struct {
	SchemaName    string  `json:"schema"`
	TableName     string  `json:"table"`
	HeapBlksRead  int64   `json:"heap_blks_read"`
	HeapBlksHit   int64   `json:"heap_blks_hit"`
	CacheHitRatio float64 `json:"cache_hit_ratio"`
	IdxBlksRead   int64   `json:"idx_blks_read"`
	IdxBlksHit    int64   `json:"idx_blks_hit"`
}

// BufferCacheInfo represents shared buffer cache statistics.
type BufferCacheInfo struct {
	TotalBuffers int64   `json:"total_buffers"`
	UsedBuffers  int64   `json:"used_buffers"`
	DirtyBuffers int64   `json:"dirty_buffers"`
	UsagePct     float64 `json:"usage_pct"`
}

const sqlWaitEvents = `
SELECT wait_event_type, wait_event, count(*) AS count
FROM pg_stat_activity
WHERE wait_event IS NOT NULL AND backend_type = 'client backend'
GROUP BY wait_event_type, wait_event
ORDER BY count DESC`

const sqlCheckpointStats = `
SELECT checkpoints_timed, checkpoints_req, buffers_checkpoint,
       buffers_clean, buffers_backend, maxwritten_clean
FROM pg_stat_bgwriter`

const sqlTableIO = `
SELECT schemaname, relname,
       heap_blks_read, heap_blks_hit,
       CASE WHEN heap_blks_read + heap_blks_hit > 0
            THEN ROUND(heap_blks_hit::numeric / (heap_blks_read + heap_blks_hit) * 100, 1)
            ELSE 100 END AS cache_hit_ratio,
       idx_blks_read, idx_blks_hit
FROM pg_statio_user_tables
WHERE schemaname = $1
ORDER BY heap_blks_read DESC`

// WaitEvents returns current wait event summary.
func WaitEvents(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]WaitEvent, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlWaitEvents)
	if err != nil {
		return nil, fmt.Errorf("pg wait events: %w", err)
	}
	defer rows.Close()
	var results []WaitEvent
	for rows.Next() {
		var e WaitEvent
		if err := rows.Scan(&e.WaitEventType, &e.WaitEvent, &e.Count); err != nil {
			return nil, fmt.Errorf("pg wait events scan: %w", err)
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

// CheckpointStats returns checkpoint statistics.
func CheckpointStats(ctx context.Context, q pginternal.Querier, _ map[string]any) (*CheckpointInfo, error) {
	row := pginternal.QueryRow(ctx, q, sqlCheckpointStats)
	var c CheckpointInfo
	if err := row.Scan(&c.CheckpointsTimed, &c.CheckpointsReq, &c.BuffersCheckpoint,
		&c.BuffersClean, &c.BuffersBackend, &c.MaxwrittenClean); err != nil {
		return nil, fmt.Errorf("pg checkpoint stats: %w", err)
	}
	return &c, nil
}

// TableIOStats returns table I/O statistics for a schema.
func TableIOStats(ctx context.Context, q pginternal.Querier, params map[string]any) ([]TableIO, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		schema = "public"
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlTableIO, schema)
	if err != nil {
		return nil, fmt.Errorf("pg table io: %w", err)
	}
	defer rows.Close()
	var results []TableIO
	for rows.Next() {
		var t TableIO
		if err := rows.Scan(&t.SchemaName, &t.TableName, &t.HeapBlksRead, &t.HeapBlksHit,
			&t.CacheHitRatio, &t.IdxBlksRead, &t.IdxBlksHit); err != nil {
			return nil, fmt.Errorf("pg table io scan: %w", err)
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

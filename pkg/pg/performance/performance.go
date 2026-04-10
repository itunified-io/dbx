// Package performance provides PostgreSQL performance analysis tools.
package performance

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// SlowQuery represents a query from pg_stat_statements.
type SlowQuery struct {
	QueryID        int64   `json:"queryid"`
	Query          string  `json:"query"`
	Calls          int64   `json:"calls"`
	TotalTime      float64 `json:"total_exec_time_ms"`
	MeanTime       float64 `json:"mean_exec_time_ms"`
	StddevTime     float64 `json:"stddev_exec_time_ms"`
	MinTime        float64 `json:"min_exec_time_ms"`
	MaxTime        float64 `json:"max_exec_time_ms"`
	Rows           int64   `json:"rows"`
	SharedBlksHit  int64   `json:"shared_blks_hit"`
	SharedBlksRead int64   `json:"shared_blks_read"`
}

// IndexAdvice represents a missing index recommendation.
type IndexAdvice struct {
	SchemaName string  `json:"schema"`
	TableName  string  `json:"table"`
	SeqScan    int64   `json:"seq_scan"`
	SeqTupRead int64   `json:"seq_tup_read"`
	IdxScan    int64   `json:"idx_scan"`
	TableSize  string  `json:"table_size"`
	SeqScanPct float64 `json:"seq_scan_pct"`
	Suggestion string  `json:"suggestion"`
}

const sqlSlowQueries = `
SELECT queryid, query, calls, total_exec_time AS total_exec_time_ms,
       mean_exec_time AS mean_exec_time_ms, stddev_exec_time AS stddev_exec_time_ms,
       min_exec_time AS min_exec_time_ms, max_exec_time AS max_exec_time_ms,
       rows, shared_blks_hit, shared_blks_read
FROM pg_stat_statements
WHERE mean_exec_time > $1
ORDER BY total_exec_time DESC
LIMIT $2`

const sqlIndexAdvisor = `
SELECT schemaname, relname,
       seq_scan, seq_tup_read, idx_scan,
       pg_size_pretty(pg_total_relation_size(quote_ident(schemaname)||'.'||quote_ident(relname))) AS table_size,
       CASE WHEN seq_scan + idx_scan > 0
            THEN ROUND(seq_scan::numeric / (seq_scan + idx_scan) * 100, 1)
            ELSE 0 END AS seq_scan_pct
FROM pg_stat_user_tables
WHERE seq_scan > idx_scan AND pg_total_relation_size(quote_ident(schemaname)||'.'||quote_ident(relname)) > 1048576
ORDER BY seq_tup_read DESC`

// SlowQueries returns queries exceeding duration threshold.
func SlowQueries(ctx context.Context, q pginternal.Querier, params map[string]any) ([]SlowQuery, error) {
	minDuration := 1000.0 // default 1 second in ms
	if v, ok := params["min_duration_ms"].(float64); ok {
		minDuration = v
	}
	limit := 20
	if v, ok := params["limit"].(int); ok {
		limit = v
	}

	rows, err := pginternal.QueryRows(ctx, q, sqlSlowQueries, minDuration, limit)
	if err != nil {
		return nil, fmt.Errorf("pg slow queries: %w", err)
	}
	defer rows.Close()
	var results []SlowQuery
	for rows.Next() {
		var sq SlowQuery
		if err := rows.Scan(&sq.QueryID, &sq.Query, &sq.Calls, &sq.TotalTime,
			&sq.MeanTime, &sq.StddevTime, &sq.MinTime, &sq.MaxTime,
			&sq.Rows, &sq.SharedBlksHit, &sq.SharedBlksRead); err != nil {
			return nil, fmt.Errorf("pg slow queries scan: %w", err)
		}
		results = append(results, sq)
	}
	return results, rows.Err()
}

// IndexAdvisor recommends missing indexes based on sequential scan ratios.
func IndexAdvisor(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]IndexAdvice, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlIndexAdvisor)
	if err != nil {
		return nil, fmt.Errorf("pg index advisor: %w", err)
	}
	defer rows.Close()
	var results []IndexAdvice
	for rows.Next() {
		var ia IndexAdvice
		if err := rows.Scan(&ia.SchemaName, &ia.TableName, &ia.SeqScan, &ia.SeqTupRead,
			&ia.IdxScan, &ia.TableSize, &ia.SeqScanPct); err != nil {
			return nil, fmt.Errorf("pg index advisor scan: %w", err)
		}
		ia.Suggestion = fmt.Sprintf("Table %s.%s has %.1f%% sequential scans (%d seq vs %d idx). Consider adding indexes.",
			ia.SchemaName, ia.TableName, ia.SeqScanPct, ia.SeqScan, ia.IdxScan)
		results = append(results, ia)
	}
	return results, rows.Err()
}

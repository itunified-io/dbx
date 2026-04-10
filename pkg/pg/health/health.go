// Package health provides PostgreSQL health monitoring tools.
package health

import (
	"context"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// AutovacuumStatus represents a table's autovacuum health.
type AutovacuumStatus struct {
	SchemaName     string     `json:"schema"`
	TableName      string     `json:"table"`
	DeadTuples     int64      `json:"n_dead_tup"`
	LastAutovacuum *time.Time `json:"last_autovacuum"`
	AutovacuumCnt  int64      `json:"autovacuum_count"`
	LiveTuples     int64      `json:"n_live_tup"`
	DeadRatio      float64    `json:"dead_ratio"`
}

// ConnectionInfo represents connection pool utilization.
type ConnectionInfo struct {
	MaxConnections      int32   `json:"max_connections"`
	CurrentConnections  int32   `json:"current_connections"`
	ReservedConnections int32   `json:"reserved_connections"`
	AvailableConns      int32   `json:"available_connections"`
	UtilizationPct      float64 `json:"utilization_pct"`
}

const sqlAutovacuumCheck = `
SELECT schemaname, relname, n_dead_tup, last_autovacuum, autovacuum_count,
       n_live_tup,
       CASE WHEN n_live_tup + n_dead_tup > 0
            THEN ROUND(n_dead_tup::numeric / (n_live_tup + n_dead_tup) * 100, 1)
            ELSE 0 END AS dead_ratio
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY n_dead_tup DESC`

const sqlConnectionLimits = `
SELECT setting::int AS max_connections,
       (SELECT count(*) FROM pg_stat_activity)::int AS current_connections,
       (SELECT setting::int FROM pg_settings WHERE name = 'superuser_reserved_connections') AS reserved_connections,
       setting::int - (SELECT count(*) FROM pg_stat_activity)::int AS available_connections,
       ROUND((SELECT count(*) FROM pg_stat_activity)::numeric / setting::numeric * 100, 1) AS utilization_pct
FROM pg_settings WHERE name = 'max_connections'`

// AutovacuumCheck returns tables with high dead tuple counts.
func AutovacuumCheck(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]AutovacuumStatus, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlAutovacuumCheck)
	if err != nil {
		return nil, fmt.Errorf("pg autovacuum check: %w", err)
	}
	defer rows.Close()
	var results []AutovacuumStatus
	for rows.Next() {
		var a AutovacuumStatus
		if err := rows.Scan(&a.SchemaName, &a.TableName, &a.DeadTuples,
			&a.LastAutovacuum, &a.AutovacuumCnt, &a.LiveTuples, &a.DeadRatio); err != nil {
			return nil, fmt.Errorf("pg autovacuum scan: %w", err)
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

// ConnectionLimits returns connection pool utilization.
func ConnectionLimits(ctx context.Context, q pginternal.Querier, _ map[string]any) (*ConnectionInfo, error) {
	row := pginternal.QueryRow(ctx, q, sqlConnectionLimits)
	var c ConnectionInfo
	if err := row.Scan(&c.MaxConnections, &c.CurrentConnections,
		&c.ReservedConnections, &c.AvailableConns, &c.UtilizationPct); err != nil {
		return nil, fmt.Errorf("pg connection limits: %w", err)
	}
	return &c, nil
}

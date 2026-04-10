// Package capacity provides PostgreSQL capacity planning tools.
package capacity

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// TableSizeInfo represents a table's storage usage.
type TableSizeInfo struct {
	SchemaName string `json:"schema"`
	TableName  string `json:"table"`
	RowCount   int64  `json:"row_count"`
	TotalSize  string `json:"total_size"`
	TableSize  string `json:"table_size"`
	IndexSize  string `json:"index_size"`
	ToastSize  string `json:"toast_size"`
}

// ForecastResult represents a storage growth forecast.
type ForecastResult struct {
	DatabaseSize    string `json:"database_size"`
	GrowthPerDay    string `json:"growth_per_day"`
	ProjectedSize30 string `json:"projected_30d"`
	ProjectedSize90 string `json:"projected_90d"`
}

const sqlTableSizes = `
SELECT schemaname, relname,
       n_live_tup AS row_count,
       pg_size_pretty(pg_total_relation_size(relid)) AS total_size,
       pg_size_pretty(pg_relation_size(relid)) AS table_size,
       pg_size_pretty(pg_indexes_size(relid)) AS index_size,
       pg_size_pretty(pg_total_relation_size(relid) - pg_indexes_size(relid) - pg_relation_size(relid)) AS toast_size
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(relid) DESC
LIMIT $1`

// TableSizes returns the largest tables by total size.
func TableSizes(ctx context.Context, q pginternal.Querier, params map[string]any) ([]TableSizeInfo, error) {
	limit := 50
	if v, ok := params["limit"].(int); ok {
		limit = v
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlTableSizes, limit)
	if err != nil {
		return nil, fmt.Errorf("pg table sizes: %w", err)
	}
	defer rows.Close()
	var results []TableSizeInfo
	for rows.Next() {
		var t TableSizeInfo
		if err := rows.Scan(&t.SchemaName, &t.TableName, &t.RowCount,
			&t.TotalSize, &t.TableSize, &t.IndexSize, &t.ToastSize); err != nil {
			return nil, fmt.Errorf("pg table sizes scan: %w", err)
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// GrowthForecast estimates storage growth based on pg_stat_user_tables deltas.
func GrowthForecast(_ context.Context, _ pginternal.Querier, _ map[string]any) (*ForecastResult, error) {
	return nil, fmt.Errorf("growth forecast requires baseline snapshot (run pg capacity baseline first)")
}

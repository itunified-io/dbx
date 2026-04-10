// Package dba_advanced provides advanced DBA operations for PostgreSQL.
package dba_advanced

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// Partition represents a table partition.
type Partition struct {
	PartitionName  string `json:"partition_name"`
	PartitionBound string `json:"partition_bound"`
	Size           string `json:"size"`
	RowEstimate    int64  `json:"row_estimate"`
}

// Matview represents a materialized view.
type Matview struct {
	SchemaName string `json:"schema"`
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	Populated  bool   `json:"populated"`
	Size       string `json:"size"`
}

// TempTable represents a temporary table.
type TempTable struct {
	SchemaName string `json:"schema"`
	TableName  string `json:"table"`
	Owner      string `json:"owner"`
	Size       string `json:"size"`
}

// AdvisoryLock represents an advisory lock.
type AdvisoryLock struct {
	LockType  string `json:"lock_type"`
	ClassID   int64  `json:"classid"`
	ObjID     int64  `json:"objid"`
	PID       int32  `json:"pid"`
	Granted   bool   `json:"granted"`
	Mode      string `json:"mode"`
}

const sqlPartitionList = `
SELECT c.relname AS partition_name, pg_get_expr(c.relpartbound, c.oid) AS partition_bound,
       pg_size_pretty(pg_total_relation_size(c.oid)) AS size,
       COALESCE(s.n_live_tup, 0) AS row_estimate
FROM pg_inherits i
JOIN pg_class c ON c.oid = i.inhrelid
JOIN pg_class p ON p.oid = i.inhparent
LEFT JOIN pg_stat_user_tables s ON s.relid = c.oid
WHERE p.relname = $1
ORDER BY c.relname`

const sqlMatviewList = `
SELECT schemaname, matviewname, matviewowner, ispopulated,
       pg_size_pretty(pg_total_relation_size(quote_ident(schemaname)||'.'||quote_ident(matviewname))) AS size
FROM pg_matviews WHERE schemaname = $1 ORDER BY matviewname`

const sqlTempTableList = `
SELECT n.nspname AS schema, c.relname AS table_name,
       pg_get_userbyid(c.relowner) AS owner,
       pg_size_pretty(pg_total_relation_size(c.oid)) AS size
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'r' AND n.nspname LIKE 'pg_temp_%'
ORDER BY c.relname`

const sqlAdvisoryLockList = `
SELECT locktype, classid, objid, pid, granted, mode
FROM pg_locks
WHERE locktype = 'advisory'
ORDER BY pid, classid, objid`

// PartitionList returns partitions of a parent table.
func PartitionList(ctx context.Context, q pginternal.Querier, params map[string]any) ([]Partition, error) {
	parent, _ := params["parent"].(string)
	if parent == "" {
		return nil, fmt.Errorf("parent parameter is required")
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlPartitionList, parent)
	if err != nil {
		return nil, fmt.Errorf("pg partition list: %w", err)
	}
	defer rows.Close()
	var results []Partition
	for rows.Next() {
		var p Partition
		if err := rows.Scan(&p.PartitionName, &p.PartitionBound, &p.Size, &p.RowEstimate); err != nil {
			return nil, fmt.Errorf("pg partition list scan: %w", err)
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

// MatviewRefresh refreshes a materialized view. Confirm-gated.
func MatviewRefresh(ctx context.Context, q pginternal.Querier, params map[string]any) (string, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return "", fmt.Errorf("confirm gate: set confirm=true to refresh materialized view")
	}
	schema, _ := params["schema"].(string)
	if schema == "" {
		schema = "public"
	}
	name, _ := params["name"].(string)
	if name == "" {
		return "", fmt.Errorf("name parameter is required")
	}
	concurrently, _ := params["concurrently"].(bool)

	sql := "REFRESH MATERIALIZED VIEW"
	if concurrently {
		sql += " CONCURRENTLY"
	}
	sql += fmt.Sprintf(" %q.%q", schema, name)

	_, err := pginternal.Exec(ctx, q, sql)
	if err != nil {
		return "", fmt.Errorf("pg matview refresh: %w", err)
	}
	return fmt.Sprintf("Materialized view %s.%s refreshed", schema, name), nil
}

// MatviewList returns materialized views in a schema.
func MatviewList(ctx context.Context, q pginternal.Querier, params map[string]any) ([]Matview, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		schema = "public"
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlMatviewList, schema)
	if err != nil {
		return nil, fmt.Errorf("pg matview list: %w", err)
	}
	defer rows.Close()
	var results []Matview
	for rows.Next() {
		var m Matview
		if err := rows.Scan(&m.SchemaName, &m.Name, &m.Owner, &m.Populated, &m.Size); err != nil {
			return nil, fmt.Errorf("pg matview list scan: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// TempTableList returns temporary tables.
func TempTableList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]TempTable, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlTempTableList)
	if err != nil {
		return nil, fmt.Errorf("pg temp table list: %w", err)
	}
	defer rows.Close()
	var results []TempTable
	for rows.Next() {
		var t TempTable
		if err := rows.Scan(&t.SchemaName, &t.TableName, &t.Owner, &t.Size); err != nil {
			return nil, fmt.Errorf("pg temp table list scan: %w", err)
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// AdvisoryLockList returns advisory locks.
func AdvisoryLockList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]AdvisoryLock, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlAdvisoryLockList)
	if err != nil {
		return nil, fmt.Errorf("pg advisory lock list: %w", err)
	}
	defer rows.Close()
	var results []AdvisoryLock
	for rows.Next() {
		var a AdvisoryLock
		if err := rows.Scan(&a.LockType, &a.ClassID, &a.ObjID, &a.PID, &a.Granted, &a.Mode); err != nil {
			return nil, fmt.Errorf("pg advisory lock list scan: %w", err)
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

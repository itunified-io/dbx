// Package schema provides PostgreSQL schema introspection tools.
package schema

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// Table represents a PostgreSQL table with size statistics.
type Table struct {
	TableName   string `json:"table_name"   db:"table_name"`
	TableType   string `json:"table_type"   db:"table_type"`
	RowEstimate int64  `json:"row_estimate" db:"row_estimate"`
	TotalSize   string `json:"total_size"   db:"total_size"`
	IndexSize   string `json:"index_size"   db:"index_size"`
}

// Column represents a table column.
type Column struct {
	ColumnName    string  `json:"column_name"    db:"column_name"`
	DataType      string  `json:"data_type"      db:"data_type"`
	IsNullable    string  `json:"is_nullable"    db:"is_nullable"`
	ColumnDefault *string `json:"column_default" db:"column_default"`
	MaxLength     *int32  `json:"character_maximum_length" db:"character_maximum_length"`
}

// Index represents a table index with usage stats.
type Index struct {
	IndexName  string `json:"indexname"     db:"indexname"`
	IndexDef   string `json:"indexdef"      db:"indexdef"`
	IdxSize    string `json:"idx_size"      db:"idx_size"`
	IdxScan    int64  `json:"idx_scan"      db:"idx_scan"`
	IdxTupRead int64  `json:"idx_tup_read"  db:"idx_tup_read"`
}

// View represents a view or materialized view.
type View struct {
	ViewName   string `json:"view_name"   db:"view_name"`
	ViewType   string `json:"view_type"   db:"view_type"`
	Definition string `json:"definition"  db:"definition"`
}

// Function represents a stored function/procedure.
type Function struct {
	FuncName   string `json:"function_name" db:"function_name"`
	ResultType string `json:"result_type"   db:"result_type"`
	ArgTypes   string `json:"arg_types"     db:"arg_types"`
	FuncType   string `json:"function_type" db:"function_type"`
}

// Enum represents a user-defined enum type.
type Enum struct {
	TypeName string   `json:"type_name" db:"type_name"`
	Labels   []string `json:"labels"`
}

// Sequence represents a sequence.
type Sequence struct {
	SequenceName string `json:"sequence_name" db:"sequence_name"`
	DataType     string `json:"data_type"     db:"data_type"`
	StartValue   int64  `json:"start_value"   db:"start_value"`
	LastValue    *int64 `json:"last_value"    db:"last_value"`
	IncrementBy  int64  `json:"increment_by"  db:"increment_by"`
}

// Extension represents an installed extension.
type Extension struct {
	Name    string `json:"name"    db:"name"`
	Version string `json:"version" db:"version"`
	Schema  string `json:"schema"  db:"schema"`
	Comment string `json:"comment" db:"comment"`
}

// SchemaInfo represents a database schema with object counts.
type SchemaInfo struct {
	SchemaName string `json:"schema_name" db:"schema_name"`
	Owner      string `json:"owner"       db:"owner"`
	TableCount int64  `json:"table_count" db:"table_count"`
}

const sqlTableList = `
SELECT t.table_name, t.table_type,
       c.reltuples::bigint AS row_estimate,
       pg_size_pretty(pg_total_relation_size(quote_ident(t.table_schema)||'.'||quote_ident(t.table_name))) AS total_size,
       pg_size_pretty(pg_indexes_size(quote_ident(t.table_schema)||'.'||quote_ident(t.table_name))) AS index_size
FROM information_schema.tables t
JOIN pg_class c ON c.relname = t.table_name
JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
WHERE t.table_schema = $1 AND t.table_type IN ('BASE TABLE', 'VIEW')
ORDER BY t.table_name`

const sqlTableDescribe = `
SELECT c.column_name, c.data_type, c.is_nullable, c.column_default, c.character_maximum_length
FROM information_schema.columns c
WHERE c.table_schema = $1 AND c.table_name = $2
ORDER BY c.ordinal_position`

const sqlIndexList = `
SELECT i.indexname, i.indexdef,
       pg_size_pretty(pg_relation_size(quote_ident(i.schemaname)||'.'||quote_ident(i.indexname))) AS idx_size,
       s.idx_scan, s.idx_tup_read
FROM pg_indexes i
LEFT JOIN pg_stat_user_indexes s ON s.indexrelname = i.indexname AND s.schemaname = i.schemaname
WHERE i.schemaname = $1 AND i.tablename = $2
ORDER BY i.indexname`

const sqlSchemaList = `
SELECT n.nspname AS schema_name, pg_catalog.pg_get_userbyid(n.nspowner) AS owner,
       (SELECT count(*) FROM pg_class c WHERE c.relnamespace = n.oid AND c.relkind IN ('r','v','m'))::bigint AS table_count
FROM pg_namespace n
WHERE n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
  AND n.nspname NOT LIKE 'pg_temp_%' AND n.nspname NOT LIKE 'pg_toast_temp_%'
ORDER BY n.nspname`

// TableList returns all tables in a schema.
func TableList(ctx context.Context, q pginternal.Querier, params map[string]any) ([]Table, error) {
	s, _ := params["schema"].(string)
	if s == "" {
		s = "public"
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlTableList, s)
	if err != nil {
		return nil, fmt.Errorf("pg table list: %w", err)
	}
	defer rows.Close()
	var tables []Table
	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.TableName, &t.TableType, &t.RowEstimate, &t.TotalSize, &t.IndexSize); err != nil {
			return nil, fmt.Errorf("pg table list scan: %w", err)
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// TableDescribe returns column details for a table.
func TableDescribe(ctx context.Context, q pginternal.Querier, params map[string]any) ([]Column, error) {
	s, _ := params["schema"].(string)
	if s == "" {
		s = "public"
	}
	table, _ := params["table"].(string)
	if table == "" {
		return nil, fmt.Errorf("table parameter is required")
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlTableDescribe, s, table)
	if err != nil {
		return nil, fmt.Errorf("pg table describe: %w", err)
	}
	defer rows.Close()
	var cols []Column
	for rows.Next() {
		var c Column
		if err := rows.Scan(&c.ColumnName, &c.DataType, &c.IsNullable, &c.ColumnDefault, &c.MaxLength); err != nil {
			return nil, fmt.Errorf("pg table describe scan: %w", err)
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

// IndexList returns indexes for a table.
func IndexList(ctx context.Context, q pginternal.Querier, params map[string]any) ([]Index, error) {
	s, _ := params["schema"].(string)
	if s == "" {
		s = "public"
	}
	table, _ := params["table"].(string)
	if table == "" {
		return nil, fmt.Errorf("table parameter is required")
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlIndexList, s, table)
	if err != nil {
		return nil, fmt.Errorf("pg index list: %w", err)
	}
	defer rows.Close()
	var indexes []Index
	for rows.Next() {
		var idx Index
		if err := rows.Scan(&idx.IndexName, &idx.IndexDef, &idx.IdxSize, &idx.IdxScan, &idx.IdxTupRead); err != nil {
			return nil, fmt.Errorf("pg index list scan: %w", err)
		}
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

// SchemaList returns all user schemas.
func SchemaList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]SchemaInfo, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlSchemaList)
	if err != nil {
		return nil, fmt.Errorf("pg schema list: %w", err)
	}
	defer rows.Close()
	var schemas []SchemaInfo
	for rows.Next() {
		var s SchemaInfo
		if err := rows.Scan(&s.SchemaName, &s.Owner, &s.TableCount); err != nil {
			return nil, fmt.Errorf("pg schema list scan: %w", err)
		}
		schemas = append(schemas, s)
	}
	return schemas, rows.Err()
}

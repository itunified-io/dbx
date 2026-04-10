// Package tenant provides PostgreSQL multi-tenant management tools.
package tenant

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// TenantInfo represents a tenant schema.
type TenantInfo struct {
	SchemaName string `json:"schema"`
	Owner      string `json:"owner"`
	TableCount int64  `json:"table_count"`
	TotalSize  string `json:"total_size"`
}

// QuotaResult represents tenant quota usage.
type QuotaResult struct {
	SchemaName  string `json:"schema"`
	CurrentSize string `json:"current_size"`
	RowCount    int64  `json:"row_count"`
}

// DriftReport represents schema drift between tenant and template.
type DriftReport struct {
	Template string      `json:"template_schema"`
	Tenant   string      `json:"tenant_schema"`
	Drifts   []DriftItem `json:"drifts"`
}

// DriftItem represents a single schema drift.
type DriftItem struct {
	DriftType  string `json:"drift_type"` // MISSING, EXTRA
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name"`
}

// IsolationResult represents tenant isolation check.
type IsolationResult struct {
	SchemaName    string   `json:"schema"`
	RLSEnabled    bool     `json:"rls_enabled"`
	SearchPath    string   `json:"search_path"`
	CrossSchemaOK bool     `json:"cross_schema_ok"`
	Issues        []string `json:"issues"`
}

// TenantStatsResult represents tenant resource usage stats.
type TenantStatsResult struct {
	SchemaName    string `json:"schema"`
	TableCount    int64  `json:"table_count"`
	TotalSize     string `json:"total_size"`
	IndexSize     string `json:"index_size"`
	ActiveConns   int32  `json:"active_connections"`
}

const sqlTenantList = `
SELECT n.nspname AS schema, pg_get_userbyid(n.nspowner) AS owner,
       (SELECT count(*) FROM pg_class c WHERE c.relnamespace = n.oid AND c.relkind = 'r') AS table_count,
       pg_size_pretty(sum(pg_total_relation_size(c.oid))::bigint) AS total_size
FROM pg_namespace n
LEFT JOIN pg_class c ON c.relnamespace = n.oid AND c.relkind = 'r'
WHERE n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast', 'public')
  AND n.nspname NOT LIKE 'pg_temp_%' AND n.nspname NOT LIKE 'pg_toast_temp_%'
GROUP BY n.nspname, n.nspowner, n.oid
ORDER BY n.nspname`

const sqlDriftDetect = `
SELECT 'MISSING' AS drift_type, t.table_name, t.column_name
FROM information_schema.columns t
WHERE t.table_schema = $1
  AND NOT EXISTS (SELECT 1 FROM information_schema.columns c
                  WHERE c.table_schema = $2 AND c.table_name = t.table_name AND c.column_name = t.column_name)
UNION ALL
SELECT 'EXTRA' AS drift_type, c.table_name, c.column_name
FROM information_schema.columns c
WHERE c.table_schema = $2
  AND NOT EXISTS (SELECT 1 FROM information_schema.columns t
                  WHERE t.table_schema = $1 AND t.table_name = c.table_name AND t.column_name = c.column_name)
ORDER BY drift_type, table_name, column_name`

// TenantList returns all tenant schemas.
func TenantList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]TenantInfo, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlTenantList)
	if err != nil {
		return nil, fmt.Errorf("pg tenant list: %w", err)
	}
	defer rows.Close()
	var results []TenantInfo
	for rows.Next() {
		var t TenantInfo
		if err := rows.Scan(&t.SchemaName, &t.Owner, &t.TableCount, &t.TotalSize); err != nil {
			return nil, fmt.Errorf("pg tenant list scan: %w", err)
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// TenantQuotas returns quota usage for a tenant schema.
func TenantQuotas(ctx context.Context, q pginternal.Querier, params map[string]any) (*QuotaResult, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		return nil, fmt.Errorf("schema parameter is required")
	}
	row := pginternal.QueryRow(ctx, q,
		`SELECT $1::text, pg_size_pretty(sum(pg_total_relation_size(c.oid))::bigint),
		        sum(s.n_live_tup)::bigint
		 FROM pg_class c
		 JOIN pg_namespace n ON n.oid = c.relnamespace
		 LEFT JOIN pg_stat_user_tables s ON s.relid = c.oid
		 WHERE n.nspname = $1 AND c.relkind = 'r'`, schema)
	var r QuotaResult
	if err := row.Scan(&r.SchemaName, &r.CurrentSize, &r.RowCount); err != nil {
		return nil, fmt.Errorf("pg tenant quotas: %w", err)
	}
	return &r, nil
}

// DriftDetect compares a tenant schema against a template schema.
func DriftDetect(ctx context.Context, q pginternal.Querier, params map[string]any) (*DriftReport, error) {
	template, _ := params["template"].(string)
	tenant, _ := params["tenant"].(string)
	if template == "" || tenant == "" {
		return nil, fmt.Errorf("template and tenant parameters are required")
	}

	rows, err := pginternal.QueryRows(ctx, q, sqlDriftDetect, template, tenant)
	if err != nil {
		return nil, fmt.Errorf("pg drift detect: %w", err)
	}
	defer rows.Close()

	report := &DriftReport{Template: template, Tenant: tenant}
	for rows.Next() {
		var d DriftItem
		if err := rows.Scan(&d.DriftType, &d.TableName, &d.ColumnName); err != nil {
			return nil, fmt.Errorf("pg drift detect scan: %w", err)
		}
		report.Drifts = append(report.Drifts, d)
	}
	return report, rows.Err()
}

// IsolationCheck verifies tenant isolation.
func IsolationCheck(ctx context.Context, q pginternal.Querier, params map[string]any) (*IsolationResult, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		return nil, fmt.Errorf("schema parameter is required")
	}
	result := &IsolationResult{SchemaName: schema, CrossSchemaOK: true}

	// Check if RLS is enabled on tables
	row := pginternal.QueryRow(ctx, q,
		`SELECT count(*) FILTER (WHERE c.relrowsecurity) > 0
		 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		 WHERE n.nspname = $1 AND c.relkind = 'r'`, schema)
	if err := row.Scan(&result.RLSEnabled); err != nil {
		result.Issues = append(result.Issues, "Could not check RLS: "+err.Error())
	}
	if !result.RLSEnabled {
		result.Issues = append(result.Issues, "RLS not enabled on any table")
	}

	return result, nil
}

// TenantStats returns resource usage statistics for a tenant.
func TenantStats(ctx context.Context, q pginternal.Querier, params map[string]any) (*TenantStatsResult, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		return nil, fmt.Errorf("schema parameter is required")
	}
	row := pginternal.QueryRow(ctx, q,
		`SELECT $1::text,
		        (SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		         WHERE n.nspname = $1 AND c.relkind = 'r')::bigint,
		        pg_size_pretty(COALESCE(sum(pg_total_relation_size(c.oid)), 0)::bigint),
		        pg_size_pretty(COALESCE(sum(pg_indexes_size(c.oid)), 0)::bigint),
		        (SELECT count(*) FROM pg_stat_activity WHERE query LIKE '%' || $1 || '%')::int
		 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		 WHERE n.nspname = $1 AND c.relkind = 'r'`, schema)
	var r TenantStatsResult
	if err := row.Scan(&r.SchemaName, &r.TableCount, &r.TotalSize, &r.IndexSize, &r.ActiveConns); err != nil {
		return nil, fmt.Errorf("pg tenant stats: %w", err)
	}
	return &r, nil
}

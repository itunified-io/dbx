// Package compliance provides PostgreSQL compliance scanning tools.
package compliance

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// ComplianceResult represents SSL compliance status.
type ComplianceResult struct {
	TotalConnections int `json:"total_connections"`
	SSLConnections   int `json:"ssl_connections"`
	NonSSLCount      int `json:"non_ssl_count"`
	Compliant        bool `json:"compliant"`
}

// PIIFinding represents a potential PII column.
type PIIFinding struct {
	SchemaName string `json:"schema"`
	TableName  string `json:"table"`
	ColumnName string `json:"column"`
	DataType   string `json:"data_type"`
	Pattern    string `json:"pattern"`
	Risk       string `json:"risk"`
}

// RetentionFinding represents a table with potential retention issues.
type RetentionFinding struct {
	SchemaName string `json:"schema"`
	TableName  string `json:"table"`
	RowCount   int64  `json:"row_count"`
	TotalSize  string `json:"total_size"`
	HasDate    bool   `json:"has_date_column"`
}

// CISReport represents a CIS benchmark scan result.
type CISReport struct {
	Passed   int          `json:"passed"`
	Failed   int          `json:"failed"`
	Warnings int          `json:"warnings"`
	Checks   []CISCheck   `json:"checks"`
}

// CISCheck represents a single CIS benchmark check.
type CISCheck struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"` // PASS, FAIL, WARN
	Detail      string `json:"detail"`
}

// FullReport combines all compliance checks.
type FullReport struct {
	SSL       *ComplianceResult  `json:"ssl"`
	PII       []PIIFinding       `json:"pii"`
	Retention []RetentionFinding `json:"retention"`
	CIS       *CISReport         `json:"cis"`
}

const sqlSSLCompliance = `
SELECT count(*) AS total,
       count(*) FILTER (WHERE ssl) AS ssl_count,
       count(*) FILTER (WHERE NOT ssl) AS non_ssl_count
FROM pg_stat_ssl s
JOIN pg_stat_activity a ON a.pid = s.pid
WHERE a.backend_type = 'client backend'`

const sqlGDPRCheck = `
SELECT table_schema, table_name, column_name, data_type
FROM information_schema.columns
WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
  AND (column_name ILIKE '%email%' OR column_name ILIKE '%phone%'
       OR column_name ILIKE '%address%' OR column_name ILIKE '%name%'
       OR column_name ILIKE '%birth%' OR column_name ILIKE '%ssn%'
       OR column_name ILIKE '%passport%' OR column_name ILIKE '%credit%')
ORDER BY table_schema, table_name, column_name`

const sqlRetentionCheck = `
SELECT s.schemaname, s.relname, s.n_live_tup,
       pg_size_pretty(pg_total_relation_size(s.relid)) AS total_size,
       EXISTS (SELECT 1 FROM information_schema.columns c
               WHERE c.table_schema = s.schemaname AND c.table_name = s.relname
               AND c.data_type IN ('timestamp without time zone', 'timestamp with time zone', 'date')) AS has_date
FROM pg_stat_user_tables s
WHERE s.n_live_tup > 10000
ORDER BY s.n_live_tup DESC`

// SSLCompliance checks SSL compliance for all connections.
func SSLCompliance(ctx context.Context, q pginternal.Querier, _ map[string]any) (*ComplianceResult, error) {
	row := pginternal.QueryRow(ctx, q, sqlSSLCompliance)
	var r ComplianceResult
	if err := row.Scan(&r.TotalConnections, &r.SSLConnections, &r.NonSSLCount); err != nil {
		return nil, fmt.Errorf("pg ssl compliance: %w", err)
	}
	r.Compliant = r.NonSSLCount == 0
	return &r, nil
}

// GDPRCheck scans for potential PII columns.
func GDPRCheck(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]PIIFinding, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlGDPRCheck)
	if err != nil {
		return nil, fmt.Errorf("pg gdpr check: %w", err)
	}
	defer rows.Close()
	var results []PIIFinding
	for rows.Next() {
		var f PIIFinding
		if err := rows.Scan(&f.SchemaName, &f.TableName, &f.ColumnName, &f.DataType); err != nil {
			return nil, fmt.Errorf("pg gdpr check scan: %w", err)
		}
		f.Pattern = classifyPIIPattern(f.ColumnName)
		f.Risk = "MEDIUM"
		results = append(results, f)
	}
	return results, rows.Err()
}

// RetentionCheck scans for tables that may need retention policies.
func RetentionCheck(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]RetentionFinding, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlRetentionCheck)
	if err != nil {
		return nil, fmt.Errorf("pg retention check: %w", err)
	}
	defer rows.Close()
	var results []RetentionFinding
	for rows.Next() {
		var f RetentionFinding
		if err := rows.Scan(&f.SchemaName, &f.TableName, &f.RowCount, &f.TotalSize, &f.HasDate); err != nil {
			return nil, fmt.Errorf("pg retention check scan: %w", err)
		}
		results = append(results, f)
	}
	return results, rows.Err()
}

// CISScan performs a basic CIS benchmark scan.
func CISScan(ctx context.Context, q pginternal.Querier, _ map[string]any) (*CISReport, error) {
	report := &CISReport{}

	// Check SSL setting
	var sslSetting string
	row := pginternal.QueryRow(ctx, q, "SELECT setting FROM pg_settings WHERE name = 'ssl'")
	if err := row.Scan(&sslSetting); err != nil {
		report.Checks = append(report.Checks, CISCheck{
			ID: "CIS-1.1", Description: "SSL enabled", Status: "WARN", Detail: "Could not check SSL setting",
		})
		report.Warnings++
	} else if sslSetting == "on" {
		report.Checks = append(report.Checks, CISCheck{
			ID: "CIS-1.1", Description: "SSL enabled", Status: "PASS", Detail: "SSL is enabled",
		})
		report.Passed++
	} else {
		report.Checks = append(report.Checks, CISCheck{
			ID: "CIS-1.1", Description: "SSL enabled", Status: "FAIL", Detail: "SSL is not enabled",
		})
		report.Failed++
	}

	return report, nil
}

// ComplianceReport generates a full compliance report.
func ComplianceReport(ctx context.Context, q pginternal.Querier, params map[string]any) (*FullReport, error) {
	report := &FullReport{}

	ssl, err := SSLCompliance(ctx, q, params)
	if err == nil {
		report.SSL = ssl
	}

	pii, err := GDPRCheck(ctx, q, params)
	if err == nil {
		report.PII = pii
	}

	retention, err := RetentionCheck(ctx, q, params)
	if err == nil {
		report.Retention = retention
	}

	cis, err := CISScan(ctx, q, params)
	if err == nil {
		report.CIS = cis
	}

	return report, nil
}

func classifyPIIPattern(columnName string) string {
	patterns := map[string]string{
		"email": "email_address", "phone": "phone_number",
		"address": "physical_address", "name": "person_name",
		"birth": "date_of_birth", "ssn": "social_security",
		"passport": "identity_document", "credit": "financial",
	}
	for key, pattern := range patterns {
		if len(columnName) >= len(key) {
			return pattern
		}
	}
	return "unknown"
}

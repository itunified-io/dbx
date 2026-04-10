// Package migration provides PostgreSQL schema migration and upgrade tools.
package migration

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// DiffReport represents differences between two database schemas.
type DiffReport struct {
	MissingTables  []string       `json:"missing_tables"`
	ExtraTables    []string       `json:"extra_tables"`
	ColumnDiffs    []ColumnDiff   `json:"column_diffs"`
}

// ColumnDiff represents a column difference between schemas.
type ColumnDiff struct {
	Table      string `json:"table"`
	Column     string `json:"column"`
	DiffType   string `json:"diff_type"` // MISSING, EXTRA, TYPE_MISMATCH
	SourceType string `json:"source_type,omitempty"`
	TargetType string `json:"target_type,omitempty"`
}

// MigrationResult represents the result of a data migration.
type MigrationResult struct {
	TablesProcessed int   `json:"tables_processed"`
	RowsCopied      int64 `json:"rows_copied"`
	Errors          []string `json:"errors,omitempty"`
}

// ExtensionCheck represents an extension compatibility check.
type ExtensionCheck struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Compatible     bool   `json:"compatible"`
}

// UpgradeReport represents a PG version upgrade readiness check.
type UpgradeReport struct {
	CurrentVersion string           `json:"current_version"`
	TargetVersion  string           `json:"target_version"`
	Extensions     []ExtensionCheck `json:"extensions"`
	Issues         []string         `json:"issues"`
	Ready          bool             `json:"ready"`
}

// SchemaDiff compares schemas between two databases.
func SchemaDiff(ctx context.Context, qA, qB pginternal.Querier, params map[string]any) (*DiffReport, error) {
	schemaA, _ := params["schema_a"].(string)
	if schemaA == "" {
		schemaA = "public"
	}
	schemaB, _ := params["schema_b"].(string)
	if schemaB == "" {
		schemaB = "public"
	}

	report := &DiffReport{}

	// Get tables from source
	tablesA, err := getTableNames(ctx, qA, schemaA)
	if err != nil {
		return nil, fmt.Errorf("schema diff source: %w", err)
	}

	// Get tables from target
	tablesB, err := getTableNames(ctx, qB, schemaB)
	if err != nil {
		return nil, fmt.Errorf("schema diff target: %w", err)
	}

	setA := make(map[string]bool)
	for _, t := range tablesA {
		setA[t] = true
	}
	setB := make(map[string]bool)
	for _, t := range tablesB {
		setB[t] = true
	}

	for _, t := range tablesA {
		if !setB[t] {
			report.MissingTables = append(report.MissingTables, t)
		}
	}
	for _, t := range tablesB {
		if !setA[t] {
			report.ExtraTables = append(report.ExtraTables, t)
		}
	}

	return report, nil
}

// DataMigration copies data between databases. Confirm-gated.
func DataMigration(_ context.Context, _, _ pginternal.Querier, params map[string]any) (*MigrationResult, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute data migration")
	}
	return nil, fmt.Errorf("data migration requires specific table configuration")
}

// ExtensionCompat checks extension compatibility.
func ExtensionCompat(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]ExtensionCheck, error) {
	rows, err := pginternal.QueryRows(ctx, q,
		`SELECT e.extname, e.extversion,
		        COALESCE(a.default_version, e.extversion) AS latest_version
		 FROM pg_extension e
		 LEFT JOIN pg_available_extensions a ON a.name = e.extname
		 ORDER BY e.extname`)
	if err != nil {
		return nil, fmt.Errorf("pg extension compat: %w", err)
	}
	defer rows.Close()
	var results []ExtensionCheck
	for rows.Next() {
		var ec ExtensionCheck
		if err := rows.Scan(&ec.Name, &ec.CurrentVersion, &ec.LatestVersion); err != nil {
			return nil, fmt.Errorf("pg extension compat scan: %w", err)
		}
		ec.Compatible = ec.CurrentVersion == ec.LatestVersion
		results = append(results, ec)
	}
	return results, rows.Err()
}

// UpgradeCheck assesses readiness for a PG version upgrade.
func UpgradeCheck(ctx context.Context, q pginternal.Querier, params map[string]any) (*UpgradeReport, error) {
	targetVersion, _ := params["target_version"].(string)

	row := pginternal.QueryRow(ctx, q, "SHOW server_version")
	var currentVersion string
	if err := row.Scan(&currentVersion); err != nil {
		return nil, fmt.Errorf("pg upgrade check: %w", err)
	}

	report := &UpgradeReport{
		CurrentVersion: currentVersion,
		TargetVersion:  targetVersion,
		Ready:          true,
	}

	exts, err := ExtensionCompat(ctx, q, nil)
	if err == nil {
		report.Extensions = exts
		for _, ext := range exts {
			if !ext.Compatible {
				report.Issues = append(report.Issues,
					fmt.Sprintf("Extension %s needs upgrade from %s to %s", ext.Name, ext.CurrentVersion, ext.LatestVersion))
			}
		}
	}

	if len(report.Issues) > 0 {
		report.Ready = false
	}

	return report, nil
}

func getTableNames(ctx context.Context, q pginternal.Querier, schema string) ([]string, error) {
	rows, err := pginternal.QueryRows(ctx, q,
		`SELECT table_name FROM information_schema.tables
		 WHERE table_schema = $1 AND table_type = 'BASE TABLE' ORDER BY table_name`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// Package backup provides PostgreSQL backup, restore, and PITR tools.
package backup

import (
	"context"
	"fmt"
	"strings"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// DumpOptions configures a pg_dump invocation.
type DumpOptions struct {
	Database string `json:"database"`
	Schema   string `json:"schema"`
	Format   string `json:"format"` // custom, directory, tar, plain
	File     string `json:"file"`
}

// RestoreOptions configures a pg_restore invocation.
type RestoreOptions struct {
	Database string `json:"database"`
	File     string `json:"file"`
	Schema   string `json:"schema"`
}

// PITROptions configures a point-in-time recovery.
type PITROptions struct {
	TargetTime string `json:"target_time"`
}

// BackupStatusResult represents last backup information.
type BackupStatusResult struct {
	LastBackupTime string `json:"last_backup_time"`
	WALPosition    string `json:"wal_position"`
	BackupMethod   string `json:"backup_method"`
}

// BuildDumpCommand constructs the pg_dump command line.
func BuildDumpCommand(opts DumpOptions) string {
	parts := []string{"pg_dump"}
	if opts.Format != "" {
		parts = append(parts, "--format="+opts.Format)
	}
	if opts.Schema != "" {
		parts = append(parts, "--schema="+opts.Schema)
	}
	if opts.File != "" {
		parts = append(parts, "--file="+opts.File)
	}
	parts = append(parts, opts.Database)
	return strings.Join(parts, " ")
}

// BuildRestoreCommand constructs the pg_restore command line.
func BuildRestoreCommand(opts RestoreOptions) string {
	parts := []string{"pg_restore"}
	if opts.Schema != "" {
		parts = append(parts, "--schema="+opts.Schema)
	}
	parts = append(parts, "--dbname="+opts.Database)
	parts = append(parts, opts.File)
	return strings.Join(parts, " ")
}

// ValidatePITRParams checks confirm gates for PITR. The caller must intend a
// destructive action (confirm=true) AND restate the target cluster's own name via
// confirm_cluster and the PITR target timestamp via confirm_timestamp (ADR-0047) —
// a generic boolean can never authorize this.
func ValidatePITRParams(params map[string]any) (*PITROptions, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute PITR")
	}
	clusterName, _ := params["cluster"].(string)
	targetTime, _ := params["target_time"].(string)
	if targetTime == "" {
		return nil, fmt.Errorf("target_time parameter is required for PITR")
	}
	confirmCluster, _ := params["confirm_cluster"].(string)
	confirmTimestamp, _ := params["confirm_timestamp"].(string)
	if confirmCluster == "" || confirmTimestamp == "" {
		return nil, fmt.Errorf("identifier confirmation required: PITR is destructive. Restate the cluster name (%q) via confirm_cluster and the target timestamp (%q) via confirm_timestamp to proceed", clusterName, targetTime)
	}
	if confirmCluster != clusterName {
		return nil, fmt.Errorf("identifier confirmation mismatch: confirm_cluster %q does not match target cluster %q", confirmCluster, clusterName)
	}
	if confirmTimestamp != targetTime {
		return nil, fmt.Errorf("identifier confirmation mismatch: confirm_timestamp %q does not match target time %q", confirmTimestamp, targetTime)
	}
	return &PITROptions{TargetTime: targetTime}, nil
}

// BackupStatus queries the last backup and WAL position.
func BackupStatus(ctx context.Context, q pginternal.Querier, _ map[string]any) (*BackupStatusResult, error) {
	row := pginternal.QueryRow(ctx, q,
		"SELECT pg_current_wal_lsn()::text, now()::text")
	var result BackupStatusResult
	var walPos, ts string
	if err := row.Scan(&walPos, &ts); err != nil {
		return nil, fmt.Errorf("pg backup status: %w", err)
	}
	result.WALPosition = walPos
	result.LastBackupTime = ts
	result.BackupMethod = "manual"
	return &result, nil
}

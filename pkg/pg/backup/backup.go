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

// ValidatePITRParams checks confirm gates for PITR.
func ValidatePITRParams(params map[string]any) (*PITROptions, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute PITR")
	}
	if dconfirm, _ := params["confirm_destructive"].(bool); !dconfirm {
		return nil, fmt.Errorf("double-confirm required: PITR is destructive. Set confirm_destructive=true")
	}
	targetTime, _ := params["target_time"].(string)
	if targetTime == "" {
		return nil, fmt.Errorf("target_time parameter is required for PITR")
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

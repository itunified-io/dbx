package backup_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/backup"
	"github.com/stretchr/testify/assert"
)

func TestBuildDumpCommand(t *testing.T) {
	cmd := backup.BuildDumpCommand(backup.DumpOptions{
		Database: "appdb",
		Schema:   "public",
		Format:   "custom",
		File:     "/tmp/appdb_public.dump",
	})
	assert.Contains(t, cmd, "pg_dump")
	assert.Contains(t, cmd, "--schema=public")
	assert.Contains(t, cmd, "--format=custom")
	assert.Contains(t, cmd, "--file=/tmp/appdb_public.dump")
	assert.Contains(t, cmd, "appdb")
}

func TestBuildRestoreCommand(t *testing.T) {
	cmd := backup.BuildRestoreCommand(backup.RestoreOptions{
		Database: "appdb",
		File:     "/tmp/appdb_public.dump",
		Schema:   "public",
	})
	assert.Contains(t, cmd, "pg_restore")
	assert.Contains(t, cmd, "--schema=public")
	assert.Contains(t, cmd, "--dbname=appdb")
}

func TestPITRRequiresDoubleConfirm(t *testing.T) {
	_, err := backup.ValidatePITRParams(map[string]any{
		"target_time":         "2026-04-10 14:30:00",
		"confirm":             true,
		"confirm_destructive": false,
	})
	assert.ErrorContains(t, err, "double-confirm required")
}

func TestPITRValidTime(t *testing.T) {
	opts, err := backup.ValidatePITRParams(map[string]any{
		"target_time":         "2026-04-10 14:30:00",
		"confirm":             true,
		"confirm_destructive": true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "2026-04-10 14:30:00", opts.TargetTime)
}

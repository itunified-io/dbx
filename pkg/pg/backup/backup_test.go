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

// ADR-0047: PITR must not proceed on a bare confirm boolean — the caller must
// restate the cluster name and target timestamp.
func TestPITRBooleanOnlyBlocks(t *testing.T) {
	_, err := backup.ValidatePITRParams(map[string]any{
		"cluster":     "pg-cluster",
		"target_time": "2026-04-10 14:30:00",
		"confirm":     true,
		// no confirm_cluster / confirm_timestamp — must block
	})
	assert.ErrorContains(t, err, "identifier confirmation required")
}

func TestPITRWrongClusterBlocks(t *testing.T) {
	_, err := backup.ValidatePITRParams(map[string]any{
		"cluster":           "pg-cluster",
		"target_time":       "2026-04-10 14:30:00",
		"confirm":           true,
		"confirm_cluster":   "wrong-cluster",
		"confirm_timestamp": "2026-04-10 14:30:00",
	})
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestPITRWrongTimestampBlocks(t *testing.T) {
	_, err := backup.ValidatePITRParams(map[string]any{
		"cluster":           "pg-cluster",
		"target_time":       "2026-04-10 14:30:00",
		"confirm":           true,
		"confirm_cluster":   "pg-cluster",
		"confirm_timestamp": "2026-01-01 00:00:00",
	})
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestPITRCorrectRestatementPassesGate(t *testing.T) {
	opts, err := backup.ValidatePITRParams(map[string]any{
		"cluster":           "pg-cluster",
		"target_time":       "2026-04-10 14:30:00",
		"confirm":           true,
		"confirm_cluster":   "pg-cluster",
		"confirm_timestamp": "2026-04-10 14:30:00",
	})
	assert.NoError(t, err)
	assert.Equal(t, "2026-04-10 14:30:00", opts.TargetTime)
}

package policy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	dir := t.TempDir()

	// Create test policy files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "security.md"), []byte("# Security Policy\nAll connections must use SSL."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "retention.md"), []byte("# Retention Policy\nData older than 90 days must be archived."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "not-a-policy.txt"), []byte("ignored"), 0644))

	store, err := policy.Status(dir)
	require.NoError(t, err)
	assert.Len(t, store.Policies, 2)
	assert.NotEmpty(t, store.Policies[0].SHA256)
}

func TestStatusEmptyDir(t *testing.T) {
	dir := t.TempDir()
	store, err := policy.Status(dir)
	require.NoError(t, err)
	assert.Empty(t, store.Policies)
}

func TestStatusMissingDir(t *testing.T) {
	_, err := policy.Status("/nonexistent/path")
	assert.Error(t, err)
}

func TestReload(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.md"), []byte("test"), 0644))

	store, err := policy.Reload(dir)
	require.NoError(t, err)
	assert.Len(t, store.Policies, 1)
}

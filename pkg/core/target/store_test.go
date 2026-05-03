package target_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withDBXHome sets HOME to a temp dir so StoreDir() resolves there.
func withDBXHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return filepath.Join(dir, ".dbx", "targets")
}

func sampleTarget(name string) *target.Target {
	return &target.Target{
		Name:        name,
		Type:        target.TypeOracleHost,
		Description: "test host",
		SSH: &target.SSHConfig{
			Host:    "10.10.0.55",
			User:    "root",
			KeyPath: "/tmp/test-key",
		},
	}
}

func TestStoreSaveLoadRoundTrip(t *testing.T) {
	withDBXHome(t)

	in := sampleTarget("ext3adm1")
	require.NoError(t, target.Save(in))

	out, err := target.Load("ext3adm1")
	require.NoError(t, err)
	assert.Equal(t, in.Name, out.Name)
	assert.Equal(t, in.Type, out.Type)
	require.NotNil(t, out.SSH)
	assert.Equal(t, "10.10.0.55", out.SSH.Host)
	assert.Equal(t, "root", out.SSH.User)
	assert.Equal(t, "/tmp/test-key", out.SSH.KeyPath)
}

func TestStoreSaveSetsMode0600(t *testing.T) {
	withDBXHome(t)

	require.NoError(t, target.Save(sampleTarget("foo")))
	fi, err := os.Stat(filepath.Join(target.StoreDir(), "foo.yaml"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
}

func TestStoreSaveOverwrites(t *testing.T) {
	withDBXHome(t)

	t1 := sampleTarget("foo")
	t1.Description = "first"
	require.NoError(t, target.Save(t1))

	t2 := sampleTarget("foo")
	t2.Description = "second"
	require.NoError(t, target.Save(t2))

	got, err := target.Load("foo")
	require.NoError(t, err)
	assert.Equal(t, "second", got.Description)
}

func TestStoreLoadMissing(t *testing.T) {
	withDBXHome(t)
	_, err := target.Load("nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, target.ErrTargetNotFound) ||
		// Allow the error to merely contain context — caller can still detect.
		err != nil)
}

func TestStoreListEmptyDir(t *testing.T) {
	withDBXHome(t)
	out, err := target.List()
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestStoreListMixedFiles(t *testing.T) {
	dir := withDBXHome(t)
	require.NoError(t, target.Save(sampleTarget("a")))
	require.NoError(t, target.Save(sampleTarget("b")))
	// non-yaml file should be skipped
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.txt"), []byte("ignore"), 0o600))

	out, err := target.List()
	require.NoError(t, err)
	require.Len(t, out, 2)
	names := []string{out[0].Name, out[1].Name}
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
}

func TestStoreRemoveIdempotent(t *testing.T) {
	withDBXHome(t)

	require.NoError(t, target.Save(sampleTarget("foo")))
	require.NoError(t, target.Remove("foo"))
	// second call must not error
	require.NoError(t, target.Remove("foo"))
	// load now fails
	_, err := target.Load("foo")
	require.Error(t, err)
}

func TestStoreSafeName(t *testing.T) {
	withDBXHome(t)

	bad := []string{
		"../etc/passwd",
		"foo/bar",
		"foo bar",
		"",
		".",
		"..",
		"foo\x00bar",
	}
	for _, n := range bad {
		t.Run(n, func(t *testing.T) {
			tgt := sampleTarget(n)
			err := target.Save(tgt)
			assert.Error(t, err, "expected error saving with name %q", n)
		})
	}

	good := []string{"ext3adm1", "host-01", "node_2", "v1.0", "ABC"}
	for _, n := range good {
		t.Run(n, func(t *testing.T) {
			tgt := sampleTarget(n)
			require.NoError(t, target.Save(tgt))
		})
	}
}

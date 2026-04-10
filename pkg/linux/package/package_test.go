package linuxpkg

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.ListArgs("oracle", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "rpm")
	assert.Contains(t, args, "-qa")
}

func TestInstallArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.InstallArgs("root", "db-host.example.com", "~/.ssh/key", "oracle-database-preinstall-19c")
	require.NoError(t, err)
	assert.Contains(t, args, "dnf")
	assert.Contains(t, args, "install")
}

func TestParseRpmList(t *testing.T) {
	raw := "kernel|5.4.17-2136.330.7.1.el8uek|x86_64\noracle-database-preinstall-19c|1.0-2.el8|x86_64\n"
	pkgs := ParseRpmList(raw)
	assert.Len(t, pkgs, 2)
	assert.Equal(t, "kernel", pkgs[0].Name)
	assert.Equal(t, "x86_64", pkgs[1].Arch)
}

func TestParseRpmList_Empty(t *testing.T) {
	pkgs := ParseRpmList("")
	assert.Empty(t, pkgs)
}

func TestIsMutating(t *testing.T) {
	assert.True(t, IsMutating("install"))
	assert.True(t, IsMutating("update"))
	assert.False(t, IsMutating("list"))
	assert.False(t, IsMutating("info"))
}

package storage

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLvListArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.LvListArgs("oracle", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "lvs")
	assert.Contains(t, args, "--noheadings")
}

func TestLvCreateArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.LvCreateArgs("root", "db-host.example.com", "~/.ssh/key", "lv_data", "50G", "vg_oracle")
	require.NoError(t, err)
	assert.Contains(t, args, "lvcreate")
	assert.Contains(t, args, "-L")
	assert.Contains(t, args, "50G")
}

func TestParseLvs(t *testing.T) {
	raw := "  lv_data|vg_oracle|100.00g\n  lv_redo|vg_oracle|20.00g\n"
	lvs := ParseLvs(raw)
	assert.Len(t, lvs, 2)
	assert.Equal(t, "lv_data", lvs[0].Name)
	assert.Equal(t, "vg_oracle", lvs[0].VG)
}

func TestParseLvs_Empty(t *testing.T) {
	lvs := ParseLvs("")
	assert.Empty(t, lvs)
}

func TestIsMutating(t *testing.T) {
	assert.True(t, IsMutating("lv_create"))
	assert.False(t, IsMutating("pv_list"))
}

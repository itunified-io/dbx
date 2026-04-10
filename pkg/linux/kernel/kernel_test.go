package kernel

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamListArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.ParamListArgs("oracle", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "sysctl")
	assert.Contains(t, args, "-a")
}

func TestHugepagesSetArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.HugepagesSetArgs("oracle", "db-host.example.com", "~/.ssh/key", 8192)
	require.NoError(t, err)
	assert.Contains(t, args, "sysctl")
	assert.Contains(t, args, "-w")
	assert.Contains(t, args, "vm.nr_hugepages=8192")
}

func TestParseSysctl(t *testing.T) {
	raw := "vm.nr_hugepages = 4096\nvm.swappiness = 10\n"
	params := ParseSysctl(raw)
	assert.Equal(t, "4096", params["vm.nr_hugepages"])
	assert.Equal(t, "10", params["vm.swappiness"])
}

func TestParseSysctl_Empty(t *testing.T) {
	params := ParseSysctl("")
	assert.Empty(t, params)
}

func TestIsMutating(t *testing.T) {
	assert.True(t, IsMutating("param_set"))
	assert.True(t, IsMutating("hugepages"))
	assert.False(t, IsMutating("param_list"))
	assert.False(t, IsMutating("info"))
}

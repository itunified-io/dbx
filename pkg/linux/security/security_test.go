package security

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelinuxStatusArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.SelinuxStatusArgs("root", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "sestatus")
}

func TestFirewallListArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.FirewallListArgs("root", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "firewall-cmd")
	assert.Contains(t, args, "--list-all")
}

func TestServiceStatusArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.ServiceStatusArgs("root", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "systemctl")
	assert.Contains(t, args, "list-units")
}

func TestP4CommandsInAllowlist(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	for _, cmd := range []string{"ip", "nmcli", "chronyc", "sestatus", "firewall-cmd", "lsblk"} {
		assert.True(t, exec.IsAllowed("linux", cmd), "missing from allowlist: %s", cmd)
	}
}

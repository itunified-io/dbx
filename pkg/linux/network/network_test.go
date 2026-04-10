package network

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNicListArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.NicListArgs("root", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "ip")
	assert.Contains(t, args, "-j")
	assert.Contains(t, args, "addr")
}

func TestNtpStatusArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.NtpStatusArgs("root", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "chronyc")
	assert.Contains(t, args, "tracking")
}

func TestDnsCheckArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.DnsCheckArgs("root", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "cat")
	assert.Contains(t, args, "/etc/resolv.conf")
}

func TestBondStatusArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	mgr := New(exec)
	args, err := mgr.BondStatusArgs("root", "db-host.example.com", "~/.ssh/key")
	require.NoError(t, err)
	assert.Contains(t, args, "nmcli")
}

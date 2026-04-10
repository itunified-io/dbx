package ssh_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/ssh"
	"github.com/stretchr/testify/assert"
)

func TestAllowlistedCommand(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	assert.True(t, exec.IsAllowed("patch", "opatch"))
	assert.True(t, exec.IsAllowed("patch", "datapatch"))
	assert.True(t, exec.IsAllowed("backup", "rman"))
	assert.True(t, exec.IsAllowed("clusterware", "crsctl"))
	assert.True(t, exec.IsAllowed("clusterware", "srvctl"))
	assert.True(t, exec.IsAllowed("asm", "asmcmd"))
	assert.True(t, exec.IsAllowed("provision", "dbca"))
	assert.True(t, exec.IsAllowed("dataguard", "dgmgrl"))
	assert.True(t, exec.IsAllowed("rac", "srvctl"))
	assert.True(t, exec.IsAllowed("linux", "rpm"))
}

func TestBlockUnknownCommand(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	assert.False(t, exec.IsAllowed("patch", "rm"))
	assert.False(t, exec.IsAllowed("backup", "sh"))
	assert.False(t, exec.IsAllowed("any", "curl"))
}

func TestBlockUnknownDomain(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	assert.False(t, exec.IsAllowed("arbitrary", "opatch"))
}

func TestBuildArgs(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	args, err := exec.BuildArgs("oracle", "db-prod.example.com", "~/.ssh/oracle_ed25519", "opatch", []string{"lspatches"})
	assert.NoError(t, err)

	assert.Equal(t, []string{
		"-i", "~/.ssh/oracle_ed25519",
		"-o", "StrictHostKeyChecking=accept-new",
		"oracle@db-prod.example.com",
		"opatch", "lspatches",
	}, args)
}

func TestBuildArgsRejectsShell(t *testing.T) {
	exec := ssh.NewExecutor(ssh.DefaultAllowlist())
	_, err := exec.BuildArgs("oracle", "host", "key", "sh", []string{"-c", "rm -rf /"})
	assert.Error(t, err)
}

func TestRegisterDomain(t *testing.T) {
	al := ssh.DefaultAllowlist()
	ssh.RegisterDomain(al, "rman-scripted", []string{"rman"})
	assert.True(t, al.Has("rman-scripted", "rman"))
}

func TestExecRequestStdinContent(t *testing.T) {
	req := ssh.ExecRequest{
		User:         "oracle",
		Host:         "db-prod.example.com",
		KeyPath:      "~/.ssh/oracle_ed25519",
		Command:      "rman",
		Args:         []string{"target", "/"},
		StdinContent: "backup database plus archivelog;\n",
	}
	assert.Equal(t, "backup database plus archivelog;\n", req.StdinContent)
}

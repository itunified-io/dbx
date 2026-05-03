package sql

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/itunified-io/dbx/pkg/host"
)

// sshExecutor implements host.Executor using the local ssh binary.
// Mirrors pkg/provision/install.sshExecutor — kept package-private here
// so the install package's executor can evolve independently (e.g.
// inventory-based credentials, sudo-flag toggling) without coupling.
type sshExecutor struct {
	target  string
	host    string
	user    string
	keyPath string
}

// newSSHExecutor returns an Executor that runs commands on the named
// dbx target via the local ssh binary. Resolves connection details from
// the dbx target registry (~/.dbx/targets/<name>.yaml). Falls back to
// the bare target name for SSH config-driven setups.
func newSSHExecutor(_ context.Context, name string) (host.Executor, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("target must not be empty")
	}
	e := &sshExecutor{target: name}
	if t, err := target.Load(name); err == nil && t.SSH != nil {
		e.host = t.SSH.Host
		e.user = t.SSH.User
		e.keyPath = t.SSH.KeyPath
	}
	return e, nil
}

func (e *sshExecutor) Run(ctx context.Context, command string) (*host.RunResult, error) {
	args := []string{
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
	}
	if e.keyPath != "" {
		args = append(args, "-i", e.keyPath)
	}
	dest := e.target
	if e.host != "" {
		if e.user != "" {
			dest = e.user + "@" + e.host
		} else {
			dest = e.host
		}
	}
	args = append(args, dest, command)
	cmd := exec.CommandContext(ctx, "ssh", args...) //nolint:gosec
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			err = nil
		} else {
			return nil, err
		}
	}
	return &host.RunResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

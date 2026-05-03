package sql

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
)

// sshExecutor implements host.Executor using the local ssh binary.
// Mirrors pkg/provision/install.sshExecutor — kept package-private here
// so the install package's executor can evolve independently (e.g.
// inventory-based credentials, sudo-flag toggling) without coupling.
type sshExecutor struct {
	target string
}

// newSSHExecutor returns an Executor that runs commands on the named
// dbx target via the local ssh binary. The target must resolve via the
// SSH config of the invoking shell (~/.ssh/config or environment); per-
// target credential resolution is the caller's responsibility (planned
// follow-up: integrate with pkg/core/target SSH endpoint resolution).
func newSSHExecutor(_ context.Context, target string) (host.Executor, error) {
	if strings.TrimSpace(target) == "" {
		return nil, fmt.Errorf("target must not be empty")
	}
	return &sshExecutor{target: target}, nil
}

func (e *sshExecutor) Run(ctx context.Context, command string) (*host.RunResult, error) {
	cmd := exec.CommandContext(ctx, "ssh", e.target, command) //nolint:gosec
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

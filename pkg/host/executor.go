// Package host defines the Executor interface used by install primitives
// to run shell commands on remote hosts.
package host

import "context"

// RunResult holds the outcome of a remote command.
type RunResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Executor runs a shell command on a remote host and returns its output.
// The concrete SSH implementation lives in pkg/host/sshexec; tests use
// pkg/host/hosttest.MockExecutor.
type Executor interface {
	Run(ctx context.Context, command string) (*RunResult, error)
}

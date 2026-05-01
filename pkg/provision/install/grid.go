package install

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
)

// GridInstall runs `runInstaller -silent` for Oracle Grid Infrastructure
// on the target host. Caller MUST have rendered + SCPed the response
// file to spec.ResponseFilePath before calling.
//
// Idempotency: if /etc/oraInst.loc and $GRID_HOME/inventory both exist
// and are populated, returns Skipped=true with Detected=Installed.
// If only one of the two is present (partial), returns an error
// pointing at --reset.
func GridInstall(ctx context.Context, spec InstallSpec, reset bool) (*InstallResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("ssh to %s: %w", spec.Target, err)
	}
	return gridInstallWithExec(ctx, exec, spec, reset)
}

// gridInstallWithExec is the testable core. Takes an injected executor
// so unit tests can use hosttest.MockExecutor.
func gridInstallWithExec(ctx context.Context, exec host.Executor, spec InstallSpec, reset bool) (*InstallResult, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	state, err := detectGridState(ctx, exec, spec.OracleHome)
	if err != nil {
		return nil, fmt.Errorf("detect grid state on %s: %w", spec.Target, err)
	}

	res := &InstallResult{Detected: state}

	switch state {
	case DetectionStateInstalled:
		if !reset {
			res.Skipped = true
			return res, nil
		}
		return nil, fmt.Errorf("--reset on installed state requires reverter primitive (not yet shipped); manual deinstall.sh required")
	case DetectionStatePartial:
		return nil, fmt.Errorf("partial install detected on %s (inventory dir incomplete); rerun with --reset after manual cleanup, or see runbook", spec.Target)
	case DetectionStateAbsent:
		// fall through to runInstaller
	}

	if spec.ResponseFilePath == "" {
		return nil, fmt.Errorf("ResponseFilePath required when state=absent (caller must render + scp the .rsp first)")
	}

	cmd := fmt.Sprintf("%s/runInstaller -silent -responseFile %s -ignorePrereqFailure",
		shellEscape(spec.OracleHome), shellEscape(spec.ResponseFilePath))
	runRes, err := exec.Run(ctx, cmd)
	if err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("runInstaller interrupted (ctx %v); remote process may still be running on %s; next run will see partial state: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("runInstaller transport failure: %w", err)
	}
	res.ExitCode = runRes.ExitCode
	res.LogTail = tailLog(runRes.Stdout+runRes.Stderr, 100)
	if runRes.ExitCode != 0 {
		return res, fmt.Errorf("runInstaller exit %d on %s; check log_tail in result", runRes.ExitCode, spec.Target)
	}
	return res, nil
}

// detectGridState probes for prior Grid install evidence per the
// Oracle 19c canonical detection rule:
//   /etc/oraInst.loc present  + $GRID_HOME/inventory non-empty → installed
//   one but not the other                                       → partial
//   neither                                                     → absent
func detectGridState(ctx context.Context, exec host.Executor, gridHome string) (DetectionState, error) {
	hasOraInst, err := probeFile(ctx, exec, "/etc/oraInst.loc")
	if err != nil {
		return DetectionStateAbsent, err
	}
	hasInventory, err := probeDirNonEmpty(ctx, exec, gridHome+"/inventory")
	if err != nil {
		return DetectionStateAbsent, err
	}
	switch {
	case hasOraInst && hasInventory:
		return DetectionStateInstalled, nil
	case !hasOraInst && !hasInventory:
		return DetectionStateAbsent, nil
	default:
		return DetectionStatePartial, nil
	}
}

// sshExecutor implements host.Executor using the local ssh binary.
// Used by GridInstall (the non-test path).
type sshExecutor struct {
	target string
}

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
			err = nil // non-zero exit is not a transport error
		} else {
			return nil, err // transport / exec failure
		}
	}
	return &host.RunResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

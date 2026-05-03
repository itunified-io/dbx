package install

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/otel"
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
func gridInstallWithExec(ctx context.Context, exec host.Executor, spec InstallSpec, reset bool) (res *InstallResult, retErr error) {
	sb := otel.NewSpan("provision.install.grid", "dbxcli").
		WithAttrs(
			otel.StringAttr(otel.AttrDbxHost, spec.Target),
			otel.StringAttr(otel.AttrDbxEntityType, "oracle_grid_home"),
			otel.StringAttr(otel.AttrDbxEntityName, spec.OracleHome),
		)
	defer func() { emitSpan(ctx, sb, retErr) }()

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	state, err := detectGridState(ctx, exec, spec.OracleHome)
	if err != nil {
		return nil, fmt.Errorf("detect grid state on %s: %w", spec.Target, err)
	}

	res = &InstallResult{Detected: state}

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

	// Oracle 19c Grid Infrastructure silent install entry-point is gridSetup.sh
	// (NOT runInstaller — that is for DB Home only). The -waitforcompletion
	// flag prevents runInstaller from forking + returning control before the
	// install actually completes (otherwise our exit code is meaningless).
	//
	// Two contextual wrappers are required:
	//   - The Grid Home is mode 0750 grid:oinstall, so gridSetup.sh MUST run
	//     as the `grid` OS user. dbx ssh logs in as root (per target.yaml),
	//     so we shell out via `sudo -u grid bash -c '...'`.
	//   - Oracle 19.3.0 base does not list OL9.x in its supported-OS table,
	//     causing INS-08101 supportedOSCheck NPE. Standard remediation is
	//     to set CV_ASSUME_DISTID to a known-good value (OEL8 covers OL9
	//     prereq parity for cluvfy + the OS check).
	cmd := fmt.Sprintf(
		"sudo -u grid -H bash -c %s",
		shellQuote(fmt.Sprintf(
			"export CV_ASSUME_DISTID=OEL8 && cd /tmp && %s/gridSetup.sh -silent -responseFile %s -ignorePrereqFailure -waitforcompletion",
			shellEscape(spec.OracleHome), shellEscape(spec.ResponseFilePath))))
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

// detectGridState probes for prior Grid install evidence.
//
// The canonical signal for Oracle 19c is /etc/oraInst.loc: it is
// created exclusively by orainstRoot.sh (run as root after the GUI
// or silent install completes) and points at the central inventory
// (oraInventory/ContentsXML/inventory.xml) where the registered
// ORACLE_HOMEs are listed.
//
// We DO NOT probe $GRID_HOME/inventory because that directory ships
// pre-populated inside the Grid Home zip (Components21/, Actions21/,
// ContentsXML/, etc.) regardless of whether the install ever ran —
// using it as a signal produces false-positive "partial" states on
// every freshly extracted Grid Home, which then forces operators to
// invoke an unimplemented --reset path. (Caught live on ext3adm1 +
// ext4adm1, infra ext3+ext4 lab, 2026-05-03.)
//
// Future enhancement: parse inventory.xml from oraInst.loc and verify
// the gridHome path is registered, distinguishing partial from full.
func detectGridState(ctx context.Context, exec host.Executor, gridHome string) (DetectionState, error) {
	_ = gridHome // reserved for future inventory.xml cross-check
	hasOraInst, err := probeFile(ctx, exec, "/etc/oraInst.loc")
	if err != nil {
		return DetectionStateAbsent, err
	}
	if hasOraInst {
		return DetectionStateInstalled, nil
	}
	return DetectionStateAbsent, nil
}

// sshExecutor implements host.Executor using the local ssh binary.
// Used by GridInstall (the non-test path) and the other install primitives
// in this package.
//
// Resolves connection details from the dbx target registry
// (~/.dbx/targets/<name>.yaml) — host, user, key_path. Falls back to
// the bare target name for SSH config-driven setups.
type sshExecutor struct {
	target  string
	host    string
	user    string
	keyPath string
}

func newSSHExecutor(_ context.Context, name string) (host.Executor, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("target must not be empty")
	}
	e := &sshExecutor{target: name}
	// Best-effort load from ~/.dbx/targets/<name>.yaml. If the load fails
	// (target not registered yet, malformed YAML, etc.) fall back to ssh
	// config defaults so existing flows keep working.
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

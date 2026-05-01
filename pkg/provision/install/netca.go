package install

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
)

// netcaSentinelDir is the directory under OracleBase where dbx writes the
// two-phase sentinel files for netca. Per package godoc, this is the
// canonical location for non-idempotent install primitives.
const netcaSentinelDir = "/cfgtoollogs/dbx"

// NetcaSilent creates an Oracle listener via `netca -silent`. Used during
// Phase D.2 (after Grid + before DBCA) to ensure a LISTENER exists for
// client connections AND during Phase E.2 to add static services
// (e.g. ORCLSTB_DGMGRL) on the standby for RMAN DUPLICATE FROM ACTIVE.
//
// Idempotency: NON-IDEMPOTENT primitive — uses the two-phase sentinel
// pattern documented in the install package godoc. Detection is
// version-agnostic (file existence + an `lsnrctl status` live probe;
// no version-string match against stdout).
//
// Reset semantics (MVP): --reset on Installed/Partial state prints a
// manual runbook to stderr and returns Skipped. Destructive listener
// drop is deferred to a reverter follow-up plan; this primitive does
// NOT delete listeners.
func NetcaSilent(ctx context.Context, spec NetcaSpec, reset bool) (*InstallResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("install: ssh to %s: %w", spec.Target, err)
	}
	return netcaSilentWithExec(ctx, exec, spec, reset)
}

// netcaSilentWithExec is the testable core. Takes an injected executor so
// unit tests can use hosttest.MockExecutor.
func netcaSilentWithExec(ctx context.Context, exec host.Executor, spec NetcaSpec, reset bool) (*InstallResult, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(spec.ResponseFilePath) == "" {
		return nil, fmt.Errorf("install: response_file_path required (caller must render + scp the netca .rsp first)")
	}

	partialPath, installedPath := netcaSentinelPaths(spec)

	state, err := detectNetcaState(ctx, exec, spec.OracleHome, spec.ListenerName, partialPath, installedPath)
	if err != nil {
		return nil, fmt.Errorf("install: detect netca state on %s: %w", spec.Target, err)
	}

	res := &InstallResult{Detected: state}

	switch state {
	case DetectionStateInstalled:
		if !reset {
			res.Skipped = true
			return res, nil
		}
		fmt.Fprint(os.Stderr, netcaResetRunbook(spec, "installed", installedPath))
		res.Skipped = true
		return res, nil
	case DetectionStatePartial:
		if !reset {
			return res, fmt.Errorf("install: partial netca state on %s (sentinel %s present without %s); rerun with --reset to print recovery runbook", spec.Target, partialPath, installedPath)
		}
		fmt.Fprint(os.Stderr, netcaResetRunbook(spec, "partial", partialPath))
		res.Skipped = true
		return res, nil
	case DetectionStateAbsent:
		// fall through
	}

	// Phase 1: write .partial sentinel BEFORE invoking netca.
	mkdirCmd := fmt.Sprintf("mkdir -p %s && : > %s",
		shellEscape(netcaSentinelRoot(spec)),
		shellEscape(partialPath),
	)
	if _, err := exec.Run(ctx, mkdirCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: netca sentinel-write interrupted (ctx %v) on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: write netca .partial sentinel on %s: %w", spec.Target, err)
	}

	// Phase 2: invoke netca -silent -responseFile <path>.
	cmd := fmt.Sprintf("%s/bin/netca -silent -responseFile %s",
		shellEscape(spec.OracleHome),
		shellEscape(spec.ResponseFilePath),
	)
	runRes, err := exec.Run(ctx, cmd)
	if err != nil {
		// Local context cancelled mid-run: remote process may still be
		// running. .partial sentinel persists so next probe sees Partial.
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: netca interrupted (ctx %v); remote process may still be running on %s; next run will see partial state: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: netca transport failure on %s: %w", spec.Target, err)
	}
	res.ExitCode = runRes.ExitCode
	res.LogTail = tailLog(runRes.Stdout+runRes.Stderr, 100)
	if runRes.ExitCode != 0 {
		// Non-zero exit: leave .partial in place so operator runs reverter.
		res.Detected = DetectionStatePartial
		return res, fmt.Errorf("install: netca exit %d on %s", runRes.ExitCode, spec.Target)
	}

	// Phase 3: atomic rename .partial → .installed.
	mvCmd := fmt.Sprintf("mv %s %s", shellEscape(partialPath), shellEscape(installedPath))
	if _, err := exec.Run(ctx, mvCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: netca sentinel-rename interrupted (ctx %v) on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: rename netca sentinel on %s: %w", spec.Target, err)
	}
	res.Detected = DetectionStateInstalled
	return res, nil
}

// detectNetcaState reads the two-phase sentinel pair PLUS a live
// `lsnrctl status <name>` probe. The live probe is included so a
// pre-existing listener (created outside dbx) is recognised as Installed
// without forcing the operator to fabricate a sentinel file.
//
//	.installed sentinel present              → Installed
//	lsnrctl status returns "STATUS of the LISTENER" → Installed
//	.partial present without .installed      → Partial
//	none of the above                        → Absent
//
// The lsnrctl probe is version-agnostic: it matches the static substring
// "STATUS of the LISTENER" present in 11g..23ai output, never a version
// number.
func detectNetcaState(ctx context.Context, exec host.Executor, oracleHome, listenerName, partialPath, installedPath string) (DetectionState, error) {
	hasInstalled, err := probeFile(ctx, exec, installedPath)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if hasInstalled {
		return DetectionStateInstalled, nil
	}
	// Live probe — a listener may exist from a prior provisioning run that
	// did not record a dbx sentinel (e.g. earlier dbx version, or manual
	// netca invocation). Treat that as Installed to avoid double-create.
	//
	// Qualify lsnrctl with $ORACLE_HOME and the absolute binary path:
	// non-interactive SSH sessions on Grid Infrastructure hosts do NOT
	// have $ORACLE_HOME/bin on $PATH, so a bare `lsnrctl` would fail with
	// "command not found" and the probe would mis-report Absent.
	probeCmd := fmt.Sprintf("env ORACLE_HOME=%s %s/bin/lsnrctl status %s",
		shellEscape(oracleHome),
		shellEscape(oracleHome),
		shellEscape(listenerName),
	)
	live, err := exec.Run(ctx, probeCmd)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if live.ExitCode == 0 && strings.Contains(live.Stdout, "STATUS of the LISTENER") {
		return DetectionStateInstalled, nil
	}
	hasPartial, err := probeFile(ctx, exec, partialPath)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if hasPartial {
		return DetectionStatePartial, nil
	}
	return DetectionStateAbsent, nil
}

// netcaSentinelRoot returns the directory under OracleBase that holds
// dbx sentinel files for this primitive.
func netcaSentinelRoot(spec NetcaSpec) string {
	return spec.OracleBase + netcaSentinelDir
}

// netcaSentinelPaths returns (partialPath, installedPath) keyed on the
// listener name (uppercased) so multiple listeners on the same host
// don't collide.
func netcaSentinelPaths(spec NetcaSpec) (string, string) {
	root := netcaSentinelRoot(spec)
	ln := strings.ToUpper(spec.ListenerName)
	return root + "/netca." + ln + ".partial",
		root + "/netca." + ln + ".installed"
}

// netcaResetRunbook returns a manual recovery runbook string that the
// caller prints to stderr when --reset is invoked. The MVP does NOT
// drop the listener; the operator must do that themselves.
func netcaResetRunbook(spec NetcaSpec, state, sentinelPath string) string {
	return fmt.Sprintf(`# netca --reset (MANUAL RUNBOOK; non-destructive in MVP)
#
# State on %s: %s
# Sentinel:   %s
# Listener:   %s (port %d)
#
# This primitive will NOT drop the listener automatically.
# Manual recovery procedure:
#
#   1. Confirm no clients depend on the listener:
#        lsnrctl services %s
#
#   2. (Operator) stop + drop the listener:
#        lsnrctl stop %s
#        # remove its entry from $ORACLE_HOME/network/admin/listener.ora
#        # (or run netca -silent -deleteListener)
#
#   3. Remove the dbx sentinel:
#        rm -f %s
#
#   4. Re-run dbxcli provision install netca (without --reset).
#
`, spec.Target, state, sentinelPath, spec.ListenerName, spec.Port,
		spec.ListenerName, spec.ListenerName, sentinelPath)
}

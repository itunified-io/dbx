package install

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/otel"
)

// dbcaSentinelDir is the directory under OracleBase where dbx writes the
// two-phase sentinel files for dbca. Per package godoc, this is the
// canonical location for non-idempotent install primitives.
const dbcaSentinelDir = "/cfgtoollogs/dbx"

// DbcaCreateDb creates an Oracle CDB via `dbca -silent -createDatabase`.
// Phase D.4 of /lab-up — runs after Grid + DB Home + listener and before
// PDB creation / Data Guard standby cloning.
//
// Idempotency: NON-IDEMPOTENT primitive — uses the two-phase sentinel
// pattern documented in the install package godoc. Detection is
// version-agnostic (file existence + a `srvctl status database -d <unique>`
// live probe; no version-string match against stdout — only the binary
// exit code is consulted).
//
// Reset semantics (MVP): --reset on Installed/Partial state prints a
// manual runbook to stderr and returns Skipped. Destructive
// `dbca -silent -deleteDatabase` is deferred to a reverter follow-up
// plan; this primitive does NOT delete databases.
func DbcaCreateDb(ctx context.Context, spec DbcaCreateDbSpec, reset bool) (*InstallResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("install: ssh to %s: %w", spec.Target, err)
	}
	return dbcaCreateDbWithExec(ctx, exec, spec, reset)
}

// dbcaCreateDbWithExec is the testable core. Takes an injected executor
// so unit tests can use hosttest.MockExecutor.
func dbcaCreateDbWithExec(ctx context.Context, exec host.Executor, spec DbcaCreateDbSpec, reset bool) (res *InstallResult, retErr error) {
	sb := otel.NewSpan("provision.install.dbca_create_db", "dbxcli").
		WithAttrs(
			otel.StringAttr(otel.AttrDbxHost, spec.Target),
			otel.StringAttr(otel.AttrDbxEntityType, "oracle_database"),
			otel.StringAttr(otel.AttrDbxEntityName, spec.DbUniqueName),
			otel.StringAttr(otel.AttrDbxDBUniqueName, spec.DbUniqueName),
		)
	defer func() { emitSpan(ctx, sb, retErr) }()

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	partialPath, installedPath := dbcaSentinelPaths(spec)

	state, err := detectDbcaState(ctx, exec, spec.OracleHome, spec.DbUniqueName, partialPath, installedPath)
	if err != nil {
		// If detection reported Partial (e.g. ctx cancelled mid-srvctl probe),
		// preserve that annotation so the caller sees Partial + error rather
		// than a bare error.
		if state == DetectionStatePartial {
			return &InstallResult{Detected: DetectionStatePartial}, fmt.Errorf("install: detect dbca state on %s: %w", spec.Target, err)
		}
		return nil, fmt.Errorf("install: detect dbca state on %s: %w", spec.Target, err)
	}

	res = &InstallResult{Detected: state}

	switch state {
	case DetectionStateInstalled:
		if !reset {
			res.Skipped = true
			return res, nil
		}
		fmt.Fprint(os.Stderr, dbcaResetRunbook(spec, "installed", installedPath))
		res.Skipped = true
		return res, nil
	case DetectionStatePartial:
		if !reset {
			return res, fmt.Errorf("install: partial dbca state on %s (sentinel %s present without %s); rerun with --reset to print recovery runbook", spec.Target, partialPath, installedPath)
		}
		fmt.Fprint(os.Stderr, dbcaResetRunbook(spec, "partial", partialPath))
		res.Skipped = true
		return res, nil
	case DetectionStateAbsent:
		// fall through
	}

	// Phase 1: write .partial sentinel BEFORE invoking dbca.
	mkdirCmd := fmt.Sprintf("mkdir -p %s && : > %s",
		shellEscape(dbcaSentinelRoot(spec)),
		shellEscape(partialPath),
	)
	if _, err := exec.Run(ctx, mkdirCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: dbca sentinel-write interrupted (ctx %v) on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: write dbca .partial sentinel on %s: %w", spec.Target, err)
	}

	// Phase 2: invoke dbca -silent -createDatabase -responseFile <path>.
	// Qualify with $ORACLE_HOME and the absolute binary path: non-
	// interactive SSH sessions on the oracle account do NOT have
	// $ORACLE_HOME/bin on $PATH, so a bare `dbca` would fail with
	// "command not found".
	cmd := fmt.Sprintf("env ORACLE_HOME=%s %s/bin/dbca -silent -createDatabase -responseFile %s",
		shellEscape(spec.OracleHome),
		shellEscape(spec.OracleHome),
		shellEscape(spec.ResponseFilePath),
	)
	runRes, err := exec.Run(ctx, cmd)
	if err != nil {
		// Local context cancelled mid-run: remote process may still be
		// running. .partial sentinel persists so next probe sees Partial.
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: dbca interrupted (ctx %v); remote process may still be running on %s; next run will see partial state: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: dbca transport failure on %s: %w", spec.Target, err)
	}
	res.ExitCode = runRes.ExitCode
	res.LogTail = tailLog(runRes.Stdout+runRes.Stderr, 100)
	if runRes.ExitCode != 0 {
		// Non-zero exit: leave .partial in place so operator runs reverter.
		res.Detected = DetectionStatePartial
		return res, fmt.Errorf("install: dbca exit %d on %s", runRes.ExitCode, spec.Target)
	}

	// Phase 3: atomic rename .partial → .installed.
	mvCmd := fmt.Sprintf("mv %s %s", shellEscape(partialPath), shellEscape(installedPath))
	if _, err := exec.Run(ctx, mvCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: dbca sentinel-rename interrupted (ctx %v) on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: rename dbca sentinel on %s: %w", spec.Target, err)
	}
	res.Detected = DetectionStateInstalled
	return res, nil
}

// detectDbcaState reads the two-phase sentinel pair PLUS a live
// `srvctl status database -d <unique>` probe. The live probe is
// included so a pre-existing database (created outside dbx) is
// recognised as Installed without forcing the operator to fabricate
// a sentinel file.
//
//	.installed sentinel present              → Installed
//	srvctl status database exit 0            → Installed
//	.partial present without .installed      → Partial
//	none of the above                        → Absent
//
// The srvctl probe is version-agnostic: only the exit code is consulted,
// never a substring match against stdout. srvctl ships with both the
// Grid Infrastructure home AND the DB home — we use the DB home path
// because dbca runs from the DB home and srvctl on that home knows
// about non-CRS-managed single-instance databases too.
func detectDbcaState(ctx context.Context, exec host.Executor, oracleHome, dbUniqueName, partialPath, installedPath string) (DetectionState, error) {
	hasInstalled, err := probeFile(ctx, exec, installedPath)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if hasInstalled {
		return DetectionStateInstalled, nil
	}
	// Live probe — qualify srvctl with $ORACLE_HOME and the absolute
	// binary path: non-interactive SSH sessions on the oracle account
	// do NOT have $ORACLE_HOME/bin on $PATH, so a bare `srvctl` would
	// fail with "command not found" and the probe would mis-report Absent.
	probeCmd := fmt.Sprintf("env ORACLE_HOME=%s %s/bin/srvctl status database -d %s",
		shellEscape(oracleHome),
		shellEscape(oracleHome),
		shellEscape(dbUniqueName),
	)
	live, err := exec.Run(ctx, probeCmd)
	if err != nil {
		// Distinguish ctx cancel (remote process may still be running →
		// Partial) from a generic transport error (DB existence unknown →
		// Absent + error is the honest answer; Partial would be a lie).
		if ctx.Err() != nil {
			return DetectionStatePartial, fmt.Errorf("srvctl probe interrupted; remote process may still be running: %w", err)
		}
		return DetectionStateAbsent, fmt.Errorf("srvctl probe transport error: %w", err)
	}
	if live.ExitCode == 0 {
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

// dbcaSentinelRoot returns the directory under OracleBase that holds
// dbx sentinel files for this primitive.
func dbcaSentinelRoot(spec DbcaCreateDbSpec) string {
	return spec.OracleBase + dbcaSentinelDir
}

// dbcaSentinelPaths returns (partialPath, installedPath) keyed on the
// DB_UNIQUE_NAME (uppercased) so multiple databases on the same host
// don't collide.
func dbcaSentinelPaths(spec DbcaCreateDbSpec) (string, string) {
	root := dbcaSentinelRoot(spec)
	u := strings.ToUpper(spec.DbUniqueName)
	return root + "/dbca." + u + ".partial",
		root + "/dbca." + u + ".installed"
}

// dbcaResetRunbook returns a manual recovery runbook string that the
// caller prints to stderr when --reset is invoked. The MVP does NOT
// drop the database; the operator must do that themselves via
// `dbca -silent -deleteDatabase` plus manual datafile cleanup.
func dbcaResetRunbook(spec DbcaCreateDbSpec, state, sentinelPath string) string {
	return fmt.Sprintf(`# dbca --reset (MANUAL RUNBOOK; non-destructive in MVP)
#
# State on %s: %s
# Sentinel:   %s
# Database:   %s (DB_UNIQUE_NAME)
#
# This primitive will NOT drop the database automatically.
# Manual recovery procedure:
#
#   1. Confirm no users / standbys depend on %s:
#        env ORACLE_HOME=%s %s/bin/srvctl status database -d %s
#        env ORACLE_HOME=%s %s/bin/sqlplus -s / as sysdba <<<'select name, open_mode from v$database;'
#
#   2. (Operator) silently delete the database:
#        env ORACLE_HOME=%s %s/bin/dbca -silent -deleteDatabase -sourceDB %s
#        # plus manual cleanup of any leftover datafiles / FRA contents
#        # under the storage destinations declared in the response file.
#
#   3. Remove the dbx sentinel:
#        rm -f %s
#
#   4. Re-run dbxcli provision install dbca (without --reset).
#
`, spec.Target, state, sentinelPath, spec.DbUniqueName,
		spec.DbUniqueName,
		spec.OracleHome, spec.OracleHome, spec.DbUniqueName,
		spec.OracleHome, spec.OracleHome,
		spec.OracleHome, spec.OracleHome, spec.DbUniqueName,
		sentinelPath)
}

package install

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
)

// PdbCreate creates an Oracle Pluggable Database via
// `dbca -silent -createPluggableDatabase`. Phase D.5 of /lab-up —
// runs after the parent CDB exists (Phase D.4 dbca -createDatabase).
//
// Idempotency: NON-IDEMPOTENT primitive — uses the two-phase sentinel
// pattern documented in the install package godoc, keyed on
// (parent CDB, new PDB) so same-PDB-name in different CDBs on the
// same host does not collide.
//
// Detection probes the .installed sentinel first, then falls back to
// a sqlplus query against v$pdbs. Detection is version-agnostic:
// only the binary exit code is consulted, never a substring match
// against stdout.
//
// Reset semantics (MVP): --reset on Installed/Partial state prints a
// manual runbook to stderr and returns Skipped. Destructive
// `dbca -silent -deletePluggableDatabase` is deferred to a reverter
// follow-up plan; this primitive does NOT drop PDBs.
func PdbCreate(ctx context.Context, spec PdbCreateSpec, reset bool) (*InstallResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("install: ssh to %s: %w", spec.Target, err)
	}
	return pdbCreateWithExec(ctx, exec, spec, reset)
}

// pdbCreateWithExec is the testable core. Takes an injected executor
// so unit tests can use hosttest.MockExecutor.
func pdbCreateWithExec(ctx context.Context, exec host.Executor, spec PdbCreateSpec, reset bool) (*InstallResult, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	partialPath, installedPath := pdbSentinelPaths(spec)

	state, err := detectPdbState(ctx, exec, spec.OracleHome, spec.PdbName, partialPath, installedPath)
	if err != nil {
		return nil, fmt.Errorf("install: detect pdb state on %s: %w", spec.Target, err)
	}

	res := &InstallResult{Detected: state}

	switch state {
	case DetectionStateInstalled:
		if !reset {
			res.Skipped = true
			return res, nil
		}
		fmt.Fprint(os.Stderr, pdbResetRunbook(spec, "installed", installedPath))
		res.Skipped = true
		return res, nil
	case DetectionStatePartial:
		if !reset {
			return res, fmt.Errorf("install: partial pdb state on %s (sentinel %s present without %s); rerun with --reset to print recovery runbook", spec.Target, partialPath, installedPath)
		}
		fmt.Fprint(os.Stderr, pdbResetRunbook(spec, "partial", partialPath))
		res.Skipped = true
		return res, nil
	case DetectionStateAbsent:
		// fall through
	}

	// Phase 1: write .partial sentinel BEFORE invoking dbca.
	mkdirCmd := fmt.Sprintf("mkdir -p %s && : > %s",
		shellEscape(pdbSentinelRoot(spec)),
		shellEscape(partialPath),
	)
	if _, err := exec.Run(ctx, mkdirCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: pdb sentinel-write interrupted (ctx %v); remote process may still be running on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: write pdb .partial sentinel on %s: %w", spec.Target, err)
	}

	// Phase 2: invoke dbca -silent -createPluggableDatabase.
	// Qualify with $ORACLE_HOME and the absolute binary path: non-
	// interactive SSH sessions on the oracle account do NOT have
	// $ORACLE_HOME/bin on $PATH, so a bare `dbca` would fail with
	// "command not found".
	//
	// Two CLI shapes are supported:
	//   - response-file form (when ResponseFilePath != ""):
	//       dbca -silent -createPluggableDatabase -responseFile <rsp>
	//   - direct-CLI form (default):
	//       dbca -silent -createPluggableDatabase -sourceDB <CDB>
	//         -pdbName <PDB> -pdbAdminUserName PDBADMIN
	//         -pdbAdminPasswordFile <file>
	//         [-pdbDatafileDestination <dest>]
	//
	// Passwords are NEVER on the command line — dbca reads
	// -pdbAdminPasswordFile from a 0600 file owned by the oracle user
	// that the caller (skill) placed before invocation.
	var dbcaCmd string
	if spec.ResponseFilePath != "" {
		dbcaCmd = fmt.Sprintf("env ORACLE_HOME=%s %s/bin/dbca -silent -createPluggableDatabase -responseFile %s",
			shellEscape(spec.OracleHome),
			shellEscape(spec.OracleHome),
			shellEscape(spec.ResponseFilePath),
		)
	} else {
		parts := []string{
			fmt.Sprintf("env ORACLE_HOME=%s %s/bin/dbca -silent -createPluggableDatabase",
				shellEscape(spec.OracleHome),
				shellEscape(spec.OracleHome),
			),
			"-sourceDB " + shellEscape(spec.CdbName),
			"-pdbName " + shellEscape(spec.PdbName),
			"-pdbAdminUserName PDBADMIN",
			"-pdbAdminPasswordFile " + shellEscape(spec.AdminPasswordFile),
		}
		if spec.DatafileDest != "" {
			parts = append(parts, "-pdbDatafileDestination "+shellEscape(spec.DatafileDest))
		}
		dbcaCmd = strings.Join(parts, " ")
	}
	runRes, err := exec.Run(ctx, dbcaCmd)
	if err != nil {
		// Local context cancelled mid-run: remote process may still be
		// running. .partial sentinel persists so next probe sees Partial.
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: dbca -createPluggableDatabase interrupted (ctx %v); remote process may still be running on %s; next run will see partial state: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: dbca -createPluggableDatabase transport failure on %s: %w", spec.Target, err)
	}
	res.ExitCode = runRes.ExitCode
	res.LogTail = tailLog(runRes.Stdout+runRes.Stderr, 100)
	if runRes.ExitCode != 0 {
		// Non-zero exit: leave .partial in place so operator runs reverter.
		res.Detected = DetectionStatePartial
		return res, fmt.Errorf("install: dbca -createPluggableDatabase exit %d on %s", runRes.ExitCode, spec.Target)
	}

	// Phase 3: atomic rename .partial → .installed.
	mvCmd := fmt.Sprintf("mv %s %s", shellEscape(partialPath), shellEscape(installedPath))
	if _, err := exec.Run(ctx, mvCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: pdb sentinel-rename interrupted (ctx %v); remote process may still be running on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: rename pdb sentinel on %s: %w", spec.Target, err)
	}
	res.Detected = DetectionStateInstalled
	return res, nil
}

// detectPdbState reads the two-phase sentinel pair PLUS a live
// `sqlplus` probe against v$pdbs.
//
//	.installed sentinel present              → Installed
//	sqlplus probe matches PDB row            → Installed
//	.partial present without .installed      → Partial
//	none of the above                        → Absent
//
// The sqlplus probe is version-agnostic: only the exit code AND
// presence of the upper-cased PDB name in stdout are consulted. The
// SQL is delivered via stdin heredoc (echo | sqlplus -s) — passing it
// on the command line would expose the query in ps(1).
func detectPdbState(ctx context.Context, exec host.Executor, oracleHome, pdbName, partialPath, installedPath string) (DetectionState, error) {
	hasInstalled, err := probeFile(ctx, exec, installedPath)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if hasInstalled {
		return DetectionStateInstalled, nil
	}
	// Live probe: select open_mode from v$pdbs where name = upper('<PDB>').
	// Empty result (no row) → exit 0 with no PDB name in stdout → not Installed.
	probeSQL := fmt.Sprintf("set heading off pagesize 0 feedback off; select name from v$pdbs where name = upper('%s'); exit;", pdbName)
	probeCmd := fmt.Sprintf("echo %s | env ORACLE_HOME=%s %s/bin/sqlplus -s / as sysdba",
		shellEscape(probeSQL),
		shellEscape(oracleHome),
		shellEscape(oracleHome),
	)
	live, err := exec.Run(ctx, probeCmd)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if live.ExitCode == 0 && strings.Contains(strings.ToUpper(live.Stdout), strings.ToUpper(pdbName)) {
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

// pdbSentinelRoot returns the directory under OracleBase that holds
// dbx sentinel files for this primitive.
func pdbSentinelRoot(spec PdbCreateSpec) string {
	// Reuse dbcaSentinelDir constant — same canonical location for all
	// non-idempotent install primitives (per package godoc).
	return spec.OracleBase + dbcaSentinelDir
}

// pdbSentinelPaths returns (partialPath, installedPath) keyed on the
// parent CDB name + PDB name (both upper-cased) so multiple PDBs on
// the same host across different CDBs don't collide.
func pdbSentinelPaths(spec PdbCreateSpec) (string, string) {
	root := pdbSentinelRoot(spec)
	cdb := strings.ToUpper(spec.CdbName)
	pdb := strings.ToUpper(spec.PdbName)
	return root + "/pdb." + cdb + "." + pdb + ".partial",
		root + "/pdb." + cdb + "." + pdb + ".installed"
}

// pdbResetRunbook returns a manual recovery runbook string. The MVP
// does NOT drop the PDB; the operator must do that themselves via
// `dbca -silent -deletePluggableDatabase`.
func pdbResetRunbook(spec PdbCreateSpec, state, sentinelPath string) string {
	return fmt.Sprintf(`# pdb create --reset (MANUAL RUNBOOK; non-destructive in MVP)
#
# State on %s: %s
# Sentinel:   %s
# Parent CDB: %s
# PDB:        %s
#
# This primitive will NOT drop the PDB automatically.
# Manual recovery procedure:
#
#   1. Confirm the PDB is closed and not in use:
#        env ORACLE_HOME=%s %s/bin/sqlplus -s / as sysdba <<<'select name, open_mode from v$pdbs;'
#
#   2. (Operator) silently delete the PDB:
#        env ORACLE_HOME=%s %s/bin/dbca -silent -deletePluggableDatabase \
#          -sourceDB %s -pdbName %s
#
#   3. Remove the dbx sentinel:
#        rm -f %s
#
#   4. Re-run dbxcli provision install pdb (without --reset).
#
`, spec.Target, state, sentinelPath, spec.CdbName, spec.PdbName,
		spec.OracleHome, spec.OracleHome,
		spec.OracleHome, spec.OracleHome,
		spec.CdbName, spec.PdbName,
		sentinelPath)
}

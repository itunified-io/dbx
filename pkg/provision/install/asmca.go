package install

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/otel"
)

// asmcaSentinelDir is the directory under OracleBase where dbx writes
// the two-phase sentinel files. Per package godoc, this is the
// canonical location for non-idempotent install primitives.
const asmcaSentinelDir = "/cfgtoollogs/dbx"

// AsmcaSilent creates an initial ASM diskgroup via `asmca -silent`.
// Used ONLY for the first DG creation during Phase D.1; subsequent
// diskgroup operations should go through mcp-oracle-ee-asm tools.
//
// Idempotency: NON-IDEMPOTENT primitive — uses the two-phase sentinel
// pattern documented in the install package godoc. Detection is
// version-agnostic (file existence only); the version routing happens
// at the caller via the OracleHome path.
//
// Reset semantics (MVP): --reset on Installed/Partial state prints a
// manual runbook to stderr and returns Skipped. Destructive
// `drop diskgroup <name> including contents` is deferred to a reverter
// follow-up plan; this primitive does NOT delete diskgroups.
func AsmcaSilent(ctx context.Context, spec AsmcaSpec, reset bool) (*InstallResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("install: ssh to %s: %w", spec.Target, err)
	}
	return asmcaSilentWithExec(ctx, exec, spec, reset)
}

// asmcaSilentWithExec is the testable core. Takes an injected executor
// so unit tests can use hosttest.MockExecutor.
func asmcaSilentWithExec(ctx context.Context, exec host.Executor, spec AsmcaSpec, reset bool) (res *InstallResult, retErr error) {
	sb := otel.NewSpan("provision.install.asmca", "dbxcli").
		WithAttrs(
			otel.StringAttr(otel.AttrDbxHost, spec.Target),
			otel.StringAttr(otel.AttrDbxEntityType, "asm_diskgroup"),
			otel.StringAttr(otel.AttrDbxEntityName, spec.DGName),
		)
	defer func() { emitSpan(ctx, sb, retErr) }()

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	partialPath, installedPath := asmcaSentinelPaths(spec)

	state, err := detectAsmcaState(ctx, exec, partialPath, installedPath)
	if err != nil {
		return nil, fmt.Errorf("install: detect asmca state on %s: %w", spec.Target, err)
	}

	res = &InstallResult{Detected: state}

	switch state {
	case DetectionStateInstalled:
		if !reset {
			res.Skipped = true
			return res, nil
		}
		fmt.Fprint(os.Stderr, asmcaResetRunbook(spec, "installed", installedPath))
		res.Skipped = true
		return res, nil
	case DetectionStatePartial:
		if !reset {
			return res, fmt.Errorf("install: partial asmca state on %s (sentinel %s present without %s); rerun with --reset to print recovery runbook", spec.Target, partialPath, installedPath)
		}
		fmt.Fprint(os.Stderr, asmcaResetRunbook(spec, "partial", partialPath))
		res.Skipped = true
		return res, nil
	case DetectionStateAbsent:
		// fall through
	}

	// Phase 1: write .partial sentinel BEFORE invoking asmca.
	mkdirCmd := fmt.Sprintf("mkdir -p %s && : > %s",
		shellEscape(asmcaSentinelRoot(spec)),
		shellEscape(partialPath),
	)
	if _, err := exec.Run(ctx, mkdirCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: asmca sentinel-write interrupted (ctx %v) on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: write asmca .partial sentinel on %s: %w", spec.Target, err)
	}

	// Phase 2: invoke asmca -silent -createDiskGroup ...
	cmd := fmt.Sprintf("%s/bin/asmca -silent -createDiskGroup -diskGroupName %s -diskList %s -redundancy %s -au_size %d",
		shellEscape(spec.OracleHome),
		shellEscape(spec.DGName),
		shellEscape(strings.Join(spec.Disks, ",")),
		spec.Redundancy,
		spec.AUSizeMB,
	)
	runRes, err := exec.Run(ctx, cmd)
	if err != nil {
		// Local context cancelled mid-run: remote process may still be
		// running. .partial sentinel persists so next probe sees Partial.
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: asmca interrupted (ctx %v); remote process may still be running on %s; next run will see partial state: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: asmca transport failure on %s: %w", spec.Target, err)
	}
	res.ExitCode = runRes.ExitCode
	res.LogTail = tailLog(runRes.Stdout+runRes.Stderr, 100)
	if runRes.ExitCode != 0 {
		// Non-zero exit: leave .partial in place so operator runs reverter.
		res.Detected = DetectionStatePartial
		return res, fmt.Errorf("install: asmca exit %d on %s", runRes.ExitCode, spec.Target)
	}

	// Phase 3: atomic rename .partial → .installed (mv is atomic on a
	// single filesystem; touch is not).
	mvCmd := fmt.Sprintf("mv %s %s", shellEscape(partialPath), shellEscape(installedPath))
	if _, err := exec.Run(ctx, mvCmd); err != nil {
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("install: asmca sentinel-rename interrupted (ctx %v) on %s: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("install: rename asmca sentinel on %s: %w", spec.Target, err)
	}
	res.Detected = DetectionStateInstalled
	return res, nil
}

// detectAsmcaState reads the two-phase sentinel pair:
//
//	.installed present                 → Installed
//	.partial present without .installed → Partial
//	neither                             → Absent
//
// Detection is version-agnostic: file existence only, no content match.
func detectAsmcaState(ctx context.Context, exec host.Executor, partialPath, installedPath string) (DetectionState, error) {
	hasInstalled, err := probeFile(ctx, exec, installedPath)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if hasInstalled {
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

// asmcaSentinelRoot returns the directory under OracleBase that holds
// dbx sentinel files for this primitive.
func asmcaSentinelRoot(spec AsmcaSpec) string {
	return spec.OracleBase + asmcaSentinelDir
}

// asmcaSentinelPaths returns (partialPath, installedPath) keyed on
// the diskgroup name so multiple DG creations on the same host don't
// collide.
func asmcaSentinelPaths(spec AsmcaSpec) (string, string) {
	root := asmcaSentinelRoot(spec)
	dg := strings.ToUpper(spec.DGName)
	return root + "/asmca." + dg + ".partial",
		root + "/asmca." + dg + ".installed"
}

// asmcaResetRunbook returns a manual recovery runbook string that the
// caller prints to stderr when --reset is invoked. The MVP does NOT
// drop the diskgroup; the operator must do that themselves.
func asmcaResetRunbook(spec AsmcaSpec, state, sentinelPath string) string {
	return fmt.Sprintf(`# asmca --reset (MANUAL RUNBOOK; non-destructive in MVP)
#
# State on %s: %s
# Sentinel:   %s
# Diskgroup:  %s
#
# This primitive will NOT drop the diskgroup automatically.
# Manual recovery procedure:
#
#   1. Confirm no databases are mounted against %s:
#        sqlplus -s / as sysasm <<<'select instance_name from v$asm_client where group_number = (select group_number from v$asm_diskgroup where name=upper('"'"'%s'"'"'));'
#
#   2. (Operator) drop the diskgroup:
#        sqlplus -s / as sysasm <<<'drop diskgroup %s including contents;'
#
#   3. Remove the dbx sentinel:
#        rm -f %s
#
#   4. Re-run dbxcli provision install asmca (without --reset).
#
`, spec.Target, state, sentinelPath, spec.DGName, spec.DGName, spec.DGName, spec.DGName, sentinelPath)
}

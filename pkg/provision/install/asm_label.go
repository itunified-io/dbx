package install

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/otel"
)

// asmLabelSentinelDir is the directory under OracleBase where dbx writes
// the per-label two-phase sentinel files. Per package godoc, this is the
// canonical location for non-idempotent install primitives.
const asmLabelSentinelDir = "/cfgtoollogs/dbx"

// oracleasmBin is the absolute path to the ASMlib admin tool. Stock
// oracleasm-support packages install it under /usr/sbin and non-
// interactive SSH sessions do NOT have /usr/sbin on $PATH for unprivileged
// users — qualifying with the absolute path mirrors the lsnrctl fix made
// to NetcaSilent (do not repeat the bare-command bug).
const oracleasmBin = "/usr/sbin/oracleasm"

// AsmDiskLabel labels raw block devices via ASMlib (oracleasm) or the
// Oracle ASM Filter Driver (AFD) so the disks become discoverable as ASM
// disks. This is a Phase D.1 prerequisite that runs BEFORE
// AsmcaSilent (which creates the diskgroup over labeled devices).
//
// Idempotency: NON-IDEMPOTENT primitive — uses the per-label two-phase
// sentinel pattern documented in the install package godoc. Each label
// gets its own .partial / .installed sentinel pair keyed on the label
// name (uppercased) so multiple disks on the same host don't collide.
// Detection is version-agnostic: file existence + a live oracleasm
// listdisks / asmcmd afd_lslbl probe; no version-string match.
//
// Reset semantics (MVP): --reset on Installed/Partial state for any
// label prints a manual recovery runbook to stderr for that label and
// records it as Skipped in the result. Destructive label removal is
// deferred to a reverter follow-up plan; this primitive does NOT delete
// labels.
//
// Per-label failure semantics: if a label fails to create (non-zero
// exit, transport failure, ctx-cancel), the function returns the
// partial AsmDiskLabelResult plus an error AND stops processing
// subsequent labels. The .partial sentinel for the failing label
// persists so a re-run sees Partial and the operator must run the
// matching reverter before retrying that label.
func AsmDiskLabel(ctx context.Context, spec AsmDiskLabelSpec, reset bool) (*AsmDiskLabelResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("install: ssh to %s: %w", spec.Target, err)
	}
	return asmDiskLabelWithExec(ctx, exec, spec, reset)
}

// asmDiskLabelWithExec is the testable core. Takes an injected executor
// so unit tests can use hosttest.MockExecutor.
func asmDiskLabelWithExec(ctx context.Context, exec host.Executor, spec AsmDiskLabelSpec, reset bool) (res *AsmDiskLabelResult, retErr error) {
	sb := otel.NewSpan("provision.install.asm_label", "dbxcli").
		WithAttrs(
			otel.StringAttr(otel.AttrDbxHost, spec.Target),
			otel.StringAttr(otel.AttrDbxEntityType, "asm_disk_labels"),
			otel.StringAttr(otel.AttrDbxEntityName, spec.Implementation),
		)
	defer func() { emitSpan(ctx, sb, retErr) }()

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	res = &AsmDiskLabelResult{Implementation: spec.Implementation}

	for _, lbl := range spec.Labels {
		partialPath, installedPath := asmLabelSentinelPaths(spec, lbl)
		lr := AsmDiskLabelLabelResult{Name: lbl.Name, Device: lbl.Device}

		state, err := detectAsmDiskLabelState(ctx, exec, spec.Implementation, spec.GridHome, lbl, partialPath, installedPath)
		if err != nil {
			lr.Detected = DetectionStateAbsent
			res.Labels = append(res.Labels, lr)
			return res, fmt.Errorf("install: detect %s label %s on %s: %w", spec.Implementation, lbl.Name, spec.Target, err)
		}
		lr.Detected = state

		switch state {
		case DetectionStateInstalled:
			if !reset {
				lr.Skipped = true
				res.Labels = append(res.Labels, lr)
				continue
			}
			fmt.Fprint(os.Stderr, asmLabelResetRunbook(spec, lbl, "installed", installedPath))
			lr.Skipped = true
			res.Labels = append(res.Labels, lr)
			continue
		case DetectionStatePartial:
			if !reset {
				res.Labels = append(res.Labels, lr)
				return res, fmt.Errorf("install: partial asm-label state on %s for %s (sentinel %s present without %s); rerun with --reset to print recovery runbook", spec.Target, lbl.Name, partialPath, installedPath)
			}
			fmt.Fprint(os.Stderr, asmLabelResetRunbook(spec, lbl, "partial", partialPath))
			lr.Skipped = true
			res.Labels = append(res.Labels, lr)
			continue
		case DetectionStateAbsent:
			// fall through
		}

		// Phase 1: write .partial sentinel BEFORE invoking the labeler.
		mkdirCmd := fmt.Sprintf("mkdir -p %s && : > %s",
			shellEscape(asmLabelSentinelRoot(spec)),
			shellEscape(partialPath),
		)
		if _, err := exec.Run(ctx, mkdirCmd); err != nil {
			if ctx.Err() != nil {
				lr.Detected = DetectionStatePartial
				res.Labels = append(res.Labels, lr)
				return res, fmt.Errorf("install: asm-label sentinel-write interrupted (ctx %v); remote process may still be running on %s for %s: %w", ctx.Err(), spec.Target, lbl.Name, err)
			}
			res.Labels = append(res.Labels, lr)
			return res, fmt.Errorf("install: write asm-label .partial sentinel for %s on %s: %w", lbl.Name, spec.Target, err)
		}

		// Phase 2: invoke the labeler.
		cmd := asmLabelCreateCmd(spec, lbl)
		runRes, err := exec.Run(ctx, cmd)
		if err != nil {
			// Local context cancelled mid-run: remote process may still be
			// running. .partial sentinel persists so next probe sees Partial.
			if ctx.Err() != nil {
				lr.Detected = DetectionStatePartial
				res.Labels = append(res.Labels, lr)
				return res, fmt.Errorf("install: asm-label %s interrupted (ctx %v); remote process may still be running on %s; next run will see partial state: %w", lbl.Name, ctx.Err(), spec.Target, err)
			}
			res.Labels = append(res.Labels, lr)
			return res, fmt.Errorf("install: asm-label transport failure on %s for %s: %w", spec.Target, lbl.Name, err)
		}
		lr.ExitCode = runRes.ExitCode
		lr.LogTail = tailLog(runRes.Stdout+runRes.Stderr, 50)
		if runRes.ExitCode != 0 {
			// Non-zero exit: leave .partial in place so operator runs reverter.
			lr.Detected = DetectionStatePartial
			res.Labels = append(res.Labels, lr)
			return res, fmt.Errorf("install: asm-label %s exit %d on %s", lbl.Name, runRes.ExitCode, spec.Target)
		}

		// Phase 3: atomic rename .partial → .installed.
		mvCmd := fmt.Sprintf("mv %s %s", shellEscape(partialPath), shellEscape(installedPath))
		if _, err := exec.Run(ctx, mvCmd); err != nil {
			if ctx.Err() != nil {
				lr.Detected = DetectionStatePartial
				res.Labels = append(res.Labels, lr)
				return res, fmt.Errorf("install: asm-label sentinel-rename interrupted (ctx %v) on %s for %s: %w", ctx.Err(), spec.Target, lbl.Name, err)
			}
			res.Labels = append(res.Labels, lr)
			return res, fmt.Errorf("install: rename asm-label sentinel for %s on %s: %w", lbl.Name, spec.Target, err)
		}
		lr.Detected = DetectionStateInstalled
		res.Labels = append(res.Labels, lr)
	}
	return res, nil
}

// asmLabelCreateCmd builds the per-impl create command. Both forms use
// absolute binary paths to defend against $PATH-less non-interactive SSH
// sessions; AFD additionally wraps with `env ORACLE_HOME=<grid_home>`
// because asmcmd reads ORACLE_HOME at startup.
func asmLabelCreateCmd(spec AsmDiskLabelSpec, lbl AsmLabelEntry) string {
	switch spec.Implementation {
	case AsmDiskLabelImplAsmlib:
		return fmt.Sprintf("%s createdisk %s %s",
			oracleasmBin,
			shellEscape(lbl.Name),
			shellEscape(lbl.Device),
		)
	case AsmDiskLabelImplAFD:
		return fmt.Sprintf("env ORACLE_HOME=%s %s/bin/asmcmd afd_label %s %s --init",
			shellEscape(spec.GridHome),
			shellEscape(spec.GridHome),
			shellEscape(lbl.Name),
			shellEscape(lbl.Device),
		)
	}
	// Validate guarantees this is unreachable; return a clearly-broken
	// command so any regression surfaces as a test failure rather than
	// silent skip.
	return fmt.Sprintf("false # install: unknown impl %q", spec.Implementation)
}

// detectAsmDiskLabelState reads the per-label two-phase sentinel pair
// PLUS a live oracleasm/afd_lslbl probe. The live probe handles labels
// that exist from a prior run that did not record a dbx sentinel.
//
//	.installed sentinel present              → Installed
//	live probe matches label                 → Installed
//	.partial present without .installed      → Partial
//	none of the above                        → Absent
//
// The live probe is version-agnostic — it only checks exit code +
// label-name presence in stdout, never matching a version string.
func detectAsmDiskLabelState(ctx context.Context, exec host.Executor, impl, gridHome string, lbl AsmLabelEntry, partialPath, installedPath string) (DetectionState, error) {
	hasInstalled, err := probeFile(ctx, exec, installedPath)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if hasInstalled {
		return DetectionStateInstalled, nil
	}
	// Live probe.
	probeCmd := asmLabelProbeCmd(impl, gridHome, lbl)
	live, err := exec.Run(ctx, probeCmd)
	if err != nil {
		return DetectionStateAbsent, err
	}
	if live.ExitCode == 0 && strings.Contains(live.Stdout, lbl.Name) {
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

// asmLabelProbeCmd builds the per-impl live probe. ASMlib uses
// `oracleasm listdisks` (lists all configured labels, one per line);
// AFD uses `asmcmd afd_lslbl <device>` which prints the label keyed
// to a specific device path. Both wrap exec invocations in absolute
// paths for the same $PATH reasons as the create command.
func asmLabelProbeCmd(impl, gridHome string, lbl AsmLabelEntry) string {
	switch impl {
	case AsmDiskLabelImplAsmlib:
		return fmt.Sprintf("%s listdisks", oracleasmBin)
	case AsmDiskLabelImplAFD:
		return fmt.Sprintf("env ORACLE_HOME=%s %s/bin/asmcmd afd_lslbl %s",
			shellEscape(gridHome),
			shellEscape(gridHome),
			shellEscape(lbl.Device),
		)
	}
	return "false"
}

// asmLabelSentinelRoot returns the directory under OracleBase that holds
// dbx sentinel files for this primitive.
func asmLabelSentinelRoot(spec AsmDiskLabelSpec) string {
	return spec.OracleBase + asmLabelSentinelDir
}

// asmLabelSentinelPaths returns (partialPath, installedPath) keyed on
// the label name (uppercased) so multiple labels on the same host
// don't collide.
func asmLabelSentinelPaths(spec AsmDiskLabelSpec, lbl AsmLabelEntry) (string, string) {
	root := asmLabelSentinelRoot(spec)
	name := strings.ToUpper(lbl.Name)
	return root + "/asm-label." + name + ".partial",
		root + "/asm-label." + name + ".installed"
}

// asmLabelResetRunbook returns a manual recovery runbook string that the
// caller prints to stderr when --reset is invoked. The MVP does NOT
// remove the label; the operator must do that themselves.
func asmLabelResetRunbook(spec AsmDiskLabelSpec, lbl AsmLabelEntry, state, sentinelPath string) string {
	var unlabel string
	switch spec.Implementation {
	case AsmDiskLabelImplAsmlib:
		unlabel = fmt.Sprintf("%s deletedisk %s", oracleasmBin, shellEscape(lbl.Name))
	case AsmDiskLabelImplAFD:
		unlabel = fmt.Sprintf("env ORACLE_HOME=%s %s/bin/asmcmd afd_unlabel %s",
			shellEscape(spec.GridHome), shellEscape(spec.GridHome), shellEscape(lbl.Name))
	}
	return fmt.Sprintf(`# asm-label --reset (MANUAL RUNBOOK; non-destructive in MVP)
#
# State on %s: %s
# Sentinel:   %s
# Label:      %s  (impl=%s, device=%s)
#
# This primitive will NOT remove the label automatically.
# Manual recovery procedure:
#
#   1. Confirm no diskgroup currently uses %s:
#        sqlplus -s / as sysasm <<<'select group_number, name from v$asm_disk where label='"'"'%s'"'"';'
#
#   2. (Operator) remove the label:
#        %s
#
#   3. Remove the dbx sentinel:
#        rm -f %s
#
#   4. Re-run dbxcli provision install asm-label (without --reset).
#
`, spec.Target, state, sentinelPath, lbl.Name, spec.Implementation, lbl.Device,
		lbl.Name, lbl.Name, unlabel, sentinelPath)
}

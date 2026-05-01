package install

import (
	"context"
	"fmt"

	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/otel"
)

// DBHomeInstall runs `runInstaller -silent` for an Oracle DB Home on
// the target host. Caller MUST have rendered + SCPed the response
// file to spec.ResponseFilePath before calling.
//
// Idempotency: if $ORACLE_HOME/bin/oracle exits 0 AND
// $ORACLE_HOME/OPatch/opatch lsinventory exits 0, returns
// Skipped=true with Detected=Installed. If only one passes,
// returns Partial + abort with --reset runbook pointer.
// Version string is not inspected — binary presence + clean exit
// is sufficient; version-routing happens at template-selection level.
func DBHomeInstall(ctx context.Context, spec InstallSpec, reset bool) (*InstallResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("ssh to %s: %w", spec.Target, err)
	}
	return dbhomeInstallWithExec(ctx, exec, spec, reset)
}

// dbhomeInstallWithExec is the testable core. Takes an injected executor
// so unit tests can use hosttest.MockExecutor.
func dbhomeInstallWithExec(ctx context.Context, exec host.Executor, spec InstallSpec, reset bool) (res *InstallResult, retErr error) {
	sb := otel.NewSpan("provision.install.dbhome", "dbxcli").
		WithAttrs(
			otel.StringAttr(otel.AttrDbxHost, spec.Target),
			otel.StringAttr(otel.AttrDbxEntityType, "oracle_db_home"),
			otel.StringAttr(otel.AttrDbxEntityName, spec.OracleHome),
		)
	defer func() { emitSpan(ctx, sb, retErr) }()

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	state, err := detectDBHomeState(ctx, exec, spec.OracleHome)
	if err != nil {
		return nil, fmt.Errorf("detect dbhome state on %s: %w", spec.Target, err)
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
		return nil, fmt.Errorf("partial dbhome install detected on %s; rerun with --reset after manual cleanup, or see runbook", spec.Target)
	case DetectionStateAbsent:
		// fall through to runInstaller
	}

	if spec.ResponseFilePath == "" {
		return nil, fmt.Errorf("ResponseFilePath required when state=absent")
	}

	cmd := fmt.Sprintf("%s/runInstaller -silent -responseFile %s -ignorePrereqFailure -waitforcompletion",
		shellEscape(spec.SoftwareStaging), shellEscape(spec.ResponseFilePath))
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
		return res, fmt.Errorf("runInstaller exit %d on %s", runRes.ExitCode, spec.Target)
	}
	return res, nil
}

// detectDBHomeState probes for prior DB home install evidence per the
// version-agnostic canonical detection rule:
//   $ORACLE_HOME/bin/oracle -V exits 0
//   AND $ORACLE_HOME/OPatch/opatch lsinventory exits 0 → installed
//   exactly one of the two                             → partial
//   neither                                            → absent
//
// The oracle binary's presence + clean exit is sufficient for all
// supported versions (19c, 21c, 23ai, 26ai). Version-routing is
// handled at the template-selection level via spec.OracleHome path.
func detectDBHomeState(ctx context.Context, exec host.Executor, oracleHome string) (DetectionState, error) {
	binCmd := fmt.Sprintf("%s/bin/oracle -V 2>&1 | head -1", shellEscape(oracleHome))
	binRes, err := exec.Run(ctx, binCmd)
	if err != nil {
		return DetectionStateAbsent, err
	}
	hasBin := binRes.ExitCode == 0

	patchCmd := fmt.Sprintf("%s/OPatch/opatch lsinventory 2>&1 | head -5", shellEscape(oracleHome))
	patchRes, err := exec.Run(ctx, patchCmd)
	if err != nil {
		return DetectionStateAbsent, err
	}
	hasOPatch := patchRes.ExitCode == 0

	switch {
	case hasBin && hasOPatch:
		return DetectionStateInstalled, nil
	case !hasBin && !hasOPatch:
		return DetectionStateAbsent, nil
	default:
		return DetectionStatePartial, nil
	}
}

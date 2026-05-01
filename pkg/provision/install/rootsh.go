package install

import (
	"context"
	"fmt"

	"github.com/itunified-io/dbx/pkg/host"
)

// RootSh runs <OracleHome>/root.sh on the target host. The wrapped
// command escalates to root via the Executor's configured ssh-user
// (must be root, or have NOPASSWD sudo). Idempotent per Oracle docs;
// we use a sentinel touchfile to skip-on-detect for clean audit
// chains. Version-agnostic: relies on file existence only, not
// content match — detection is valid for 19c, 21c, 23ai, and 26ai.
//
// Unlike Grid/DBHome, --reset is allowed on this primitive: root.sh
// is documented as safe to re-run, so --reset simply re-executes
// without erroring.
func RootSh(ctx context.Context, spec InstallSpec, reset bool) (*InstallResult, error) {
	exec, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("ssh to %s: %w", spec.Target, err)
	}
	return rootShWithExec(ctx, exec, spec, reset)
}

// rootShWithExec is the testable core. Takes an injected executor so
// unit tests can use hosttest.MockExecutor.
func rootShWithExec(ctx context.Context, exec host.Executor, spec InstallSpec, reset bool) (*InstallResult, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	touch := spec.OracleHome + "/install/rootsh.touchfile"
	exists, err := probeFile(ctx, exec, touch)
	if err != nil {
		return nil, fmt.Errorf("probe touchfile on %s: %w", spec.Target, err)
	}

	res := &InstallResult{}
	if exists {
		res.Detected = DetectionStateInstalled
		if !reset {
			res.Skipped = true
			return res, nil
		}
		// fall through: --reset means "re-run anyway" (root.sh is idempotent)
	} else {
		res.Detected = DetectionStateAbsent
	}

	cmd := fmt.Sprintf("%s/root.sh && touch %s", shellEscape(spec.OracleHome), shellEscape(touch))
	runRes, err := exec.Run(ctx, cmd)
	if err != nil {
		// If the local context was cancelled, the remote process may
		// still be running on the target. Surface this as Partial so
		// the next probe can pick it up; touchfile won't exist yet.
		if ctx.Err() != nil {
			res.Detected = DetectionStatePartial
			return res, fmt.Errorf("root.sh interrupted (ctx %v); remote process may still be running on %s; next run will see partial state: %w", ctx.Err(), spec.Target, err)
		}
		return nil, fmt.Errorf("root.sh transport failure: %w", err)
	}
	res.ExitCode = runRes.ExitCode
	res.LogTail = tailLog(runRes.Stdout+runRes.Stderr, 100)
	if runRes.ExitCode != 0 {
		return res, fmt.Errorf("install: root.sh exit %d on %s", runRes.ExitCode, spec.Target)
	}
	return res, nil
}

package install

import (
	"context"
	"errors"
	"testing"

	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/host/hosttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootSh(t *testing.T) {
	cases := []struct {
		name            string
		setupMock       func(*hosttest.MockExecutor)
		spec            InstallSpec
		reset           bool
		wantErr         string
		wantSkipped     bool
		wantDetected    DetectionState
		wantLogContains string
	}{
		{
			name: "DetectsExistingTouchfile_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(0, "ran 2026-04-30T10:00:00Z\n", "")
			},
			spec:         InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"},
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "AbsentTouchfile_RunsScript",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(1, "", "")
				m.OnCommand("/u01/app/19c/grid/root.sh && touch /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(0, "Performing root user operation.\nThe following environment variables are set as:\n\n", "")
			},
			spec:            InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"},
			wantSkipped:     false,
			wantDetected:    DetectionStateAbsent,
			wantLogContains: "Performing root user operation",
		},
		{
			name: "ScriptExitNonZero_ReturnsError",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(1, "", "")
				m.OnCommand("/u01/app/19c/grid/root.sh && touch /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(5, "ERROR: ROOT.SH operation failed\n", "")
			},
			spec:            InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"},
			wantErr:         "exit 5",
			wantLogContains: "ROOT.SH operation failed",
			wantDetected:    DetectionStateAbsent,
		},
		{
			name:      "RejectsInvalidSpec",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec:      InstallSpec{},
			wantErr:   "target is required",
		},
		{
			name: "ResetOnInstalled_ReRunsScript",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(0, "ran 2026-04-30T10:00:00Z\n", "")
				m.OnCommand("/u01/app/19c/grid/root.sh && touch /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(0, "Performing root user operation again.\n", "")
			},
			spec:            InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"},
			reset:           true,
			wantSkipped:     false,
			wantDetected:    DetectionStateInstalled,
			wantLogContains: "again",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := rootShWithExec(context.Background(), mock, tc.spec, tc.reset)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				if tc.wantLogContains != "" && res != nil {
					assert.Contains(t, res.LogTail, tc.wantLogContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tc.wantSkipped, res.Skipped)
			assert.Equal(t, tc.wantDetected, res.Detected)
			if tc.wantLogContains != "" {
				assert.Contains(t, res.LogTail, tc.wantLogContains)
			}
		})
	}
}

// ctxCancelExec runs the first command (the touchfile probe) normally
// via the inner mock; on the SECOND command (the install command) it
// behaves as if the context had been cancelled mid-run — it returns
// ctx.Err() (or a sentinel) without ever producing a RunResult. The
// caller's ctx is also cancelled before the second call to make
// ctx.Err() reflect that state.
type ctxCancelExec struct {
	inner   host.Executor
	calls   int
	cancel  context.CancelFunc
	wrapErr error
}

func (e *ctxCancelExec) Run(ctx context.Context, cmd string) (*host.RunResult, error) {
	e.calls++
	if e.calls == 1 {
		return e.inner.Run(ctx, cmd)
	}
	// Simulate ctx cancellation observed by the underlying transport.
	if e.cancel != nil {
		e.cancel()
	}
	return nil, e.wrapErr
}

func TestRootSh_CtxCancelled_ReportsPartial(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile").Returns(1, "", "")

	ctx, cancel := context.WithCancel(context.Background())
	ex := &ctxCancelExec{
		inner:   mock,
		cancel:  cancel,
		wrapErr: errors.New("transport: connection closed"),
	}

	spec := InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"}
	res, err := rootShWithExec(ctx, ex, spec, false)

	require.Error(t, err)
	require.NotNil(t, res)
	assert.Equal(t, DetectionStatePartial, res.Detected)
	assert.Contains(t, err.Error(), "interrupted")
	assert.Contains(t, err.Error(), "may still be running")
}

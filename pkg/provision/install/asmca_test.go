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

// validAsmcaSpec returns a minimal valid spec usable across tests.
func validAsmcaSpec() AsmcaSpec {
	return AsmcaSpec{
		InstallSpec: InstallSpec{
			Target:     "ext3adm1",
			OracleHome: "/u01/app/19c/grid",
			OracleBase: "/u01/app/grid",
		},
		DGName:     "DATA",
		Redundancy: "EXTERNAL",
		AUSizeMB:   4,
		Disks:      []string{"/dev/sdb", "/dev/sdc"},
	}
}

func TestAsmcaSilent(t *testing.T) {
	const partial = "/u01/app/grid/cfgtoollogs/dbx/asmca.DATA.partial"
	const installed = "/u01/app/grid/cfgtoollogs/dbx/asmca.DATA.installed"
	const mkdir = "mkdir -p /u01/app/grid/cfgtoollogs/dbx && : > /u01/app/grid/cfgtoollogs/dbx/asmca.DATA.partial"
	const asmca = "/u01/app/19c/grid/bin/asmca -silent -createDiskGroup -diskGroupName DATA -diskList /dev/sdb,/dev/sdc -redundancy EXTERNAL -au_size 4"
	const mv = "mv /u01/app/grid/cfgtoollogs/dbx/asmca.DATA.partial /u01/app/grid/cfgtoollogs/dbx/asmca.DATA.installed"

	cases := []struct {
		name            string
		setupMock       func(*hosttest.MockExecutor)
		spec            AsmcaSpec
		reset           bool
		wantErr         string
		wantSkipped     bool
		wantDetected    DetectionState
		wantLogContains string
	}{
		{
			name: "InstalledSentinelDetected_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(0, "", "")
			},
			spec:         validAsmcaSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "PartialSentinelDetected_ReturnsError",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validAsmcaSpec(),
			wantErr:      "partial asmca state",
			wantDetected: DetectionStatePartial,
		},
		{
			name: "AbsentSentinel_RunsFullPipeline",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(asmca).Returns(0, "Disk Group DATA created successfully.\n", "")
				m.OnCommand(mv).Returns(0, "", "")
			},
			spec:            validAsmcaSpec(),
			wantSkipped:     false,
			wantDetected:    DetectionStateInstalled,
			wantLogContains: "created successfully",
		},
		{
			name: "AsmcaExitNonzero_LeavesPartialAndErrors",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(asmca).Returns(1, "ERROR: ASMCA-08105 Insufficient disks\n", "")
			},
			spec:            validAsmcaSpec(),
			wantErr:         "exit 1",
			wantDetected:    DetectionStatePartial,
			wantLogContains: "ASMCA-08105",
		},
		{
			name: "Reset_PrintsRunbook_OnInstalled_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(0, "", "")
			},
			spec:         validAsmcaSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "Reset_OnPartial_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validAsmcaSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStatePartial,
		},
		{
			name:      "RejectsMissingTarget",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmcaSpec{
				InstallSpec: InstallSpec{OracleHome: "/x", OracleBase: "/y"},
				DGName:      "D", Redundancy: "EXTERNAL", AUSizeMB: 4, Disks: []string{"/d"},
			},
			wantErr: "target is required",
		},
		{
			name:      "RejectsMissingOracleBase",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmcaSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x"},
				DGName:      "D", Redundancy: "EXTERNAL", AUSizeMB: 4, Disks: []string{"/d"},
			},
			wantErr: "oracle_base is required",
		},
		{
			name:      "RejectsMissingDGName",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmcaSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				Redundancy:  "EXTERNAL", AUSizeMB: 4, Disks: []string{"/d"},
			},
			wantErr: "dg_name is required",
		},
		{
			name:      "RejectsBadRedundancy",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmcaSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				DGName:      "D", Redundancy: "BOGUS", AUSizeMB: 4, Disks: []string{"/d"},
			},
			wantErr: "EXTERNAL/NORMAL/HIGH",
		},
		{
			name:      "RejectsZeroAUSize",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmcaSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				DGName:      "D", Redundancy: "EXTERNAL", AUSizeMB: 0, Disks: []string{"/d"},
			},
			wantErr: "au_size_mb",
		},
		{
			name:      "RejectsEmptyDisks",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmcaSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				DGName:      "D", Redundancy: "EXTERNAL", AUSizeMB: 4,
			},
			wantErr: "disks list is required",
		},
		{
			name:      "RejectsCommaInDisk",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmcaSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				DGName:      "D", Redundancy: "EXTERNAL", AUSizeMB: 4,
				Disks: []string{"/dev/sdb,/dev/sdc"},
			},
			wantErr: "control character or comma",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := asmcaSilentWithExec(context.Background(), mock, tc.spec, tc.reset)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				if tc.wantLogContains != "" && res != nil {
					assert.Contains(t, res.LogTail, tc.wantLogContains)
				}
				if tc.wantDetected != 0 && res != nil {
					assert.Equal(t, tc.wantDetected, res.Detected)
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

// asmcaCtxCancelExec mimics rootsh_test.go's helper: lets the first N
// probe calls go to the inner mock, then cancels ctx and fails the
// (N+1)-th call to simulate a transport disconnect mid-run.
type asmcaCtxCancelExec struct {
	inner       host.Executor
	cancelAfter int // number of inner calls before cancelling
	calls       int
	cancel      context.CancelFunc
	wrapErr     error
}

func (e *asmcaCtxCancelExec) Run(ctx context.Context, cmd string) (*host.RunResult, error) {
	e.calls++
	if e.calls <= e.cancelAfter {
		return e.inner.Run(ctx, cmd)
	}
	if e.cancel != nil {
		e.cancel()
	}
	return nil, e.wrapErr
}

func TestAsmcaSilent_CtxCancelled_ReportsPartial(t *testing.T) {
	const partial = "/u01/app/grid/cfgtoollogs/dbx/asmca.DATA.partial"
	const installed = "/u01/app/grid/cfgtoollogs/dbx/asmca.DATA.installed"
	const mkdir = "mkdir -p /u01/app/grid/cfgtoollogs/dbx && : > /u01/app/grid/cfgtoollogs/dbx/asmca.DATA.partial"

	mock := hosttest.NewMockExecutor()
	// detection: both probes report absent
	mock.OnCommand("test -f " + installed).Returns(1, "", "")
	mock.OnCommand("test -f " + partial).Returns(1, "", "")
	// sentinel write: succeeds (so .partial exists on the host)
	mock.OnCommand(mkdir).Returns(0, "", "")
	// asmca call is intercepted and ctx-cancelled by the wrapper

	ctx, cancel := context.WithCancel(context.Background())
	ex := &asmcaCtxCancelExec{
		inner:       mock,
		cancelAfter: 3, // 2 probes + 1 sentinel-write through the inner mock
		cancel:      cancel,
		wrapErr:     errors.New("transport: connection closed"),
	}

	res, err := asmcaSilentWithExec(ctx, ex, validAsmcaSpec(), false)
	require.Error(t, err)
	require.NotNil(t, res)
	assert.Equal(t, DetectionStatePartial, res.Detected)
	assert.Contains(t, err.Error(), "interrupted")
	assert.Contains(t, err.Error(), "may still be running")
}

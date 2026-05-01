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

// validNetcaSpec returns a minimal valid spec usable across tests.
func validNetcaSpec() NetcaSpec {
	return NetcaSpec{
		InstallSpec: InstallSpec{
			Target:           "ext3adm1",
			OracleHome:       "/u01/app/oracle/product/19c/dbhome_1",
			OracleBase:       "/u01/app/oracle",
			ResponseFilePath: "/tmp/netca.rsp",
		},
		ListenerName: "LISTENER",
		Port:         1521,
	}
}

func TestNetcaSilent(t *testing.T) {
	const partial = "/u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.partial"
	const installed = "/u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.installed"
	const lsnrctl = "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/lsnrctl status LISTENER"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.partial"
	const netca = "/u01/app/oracle/product/19c/dbhome_1/bin/netca -silent -responseFile /tmp/netca.rsp"
	const mv = "mv /u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.partial /u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.installed"

	const lsnrctlOk = "LSNRCTL for Linux: Version 19.0.0.0.0\n\nSTATUS of the LISTENER\n------------------------\nAlias                     LISTENER\n"
	const lsnrctlAbsent = "TNS-12541: TNS:no listener\n"

	cases := []struct {
		name            string
		setupMock       func(*hosttest.MockExecutor)
		spec            NetcaSpec
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
			spec:         validNetcaSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "LiveListenerDetected_NoSentinel_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(lsnrctl).Returns(0, lsnrctlOk, "")
			},
			spec:         validNetcaSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "PartialSentinelDetected_ReturnsError",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(lsnrctl).Returns(1, lsnrctlAbsent, "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validNetcaSpec(),
			wantErr:      "partial netca state",
			wantDetected: DetectionStatePartial,
		},
		{
			name: "AbsentListener_RunsFullPipeline",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(lsnrctl).Returns(1, lsnrctlAbsent, "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(netca).Returns(0, "Oracle Net Listener Startup:\nLISTENER successfully started.\n", "")
				m.OnCommand(mv).Returns(0, "", "")
			},
			spec:            validNetcaSpec(),
			wantSkipped:     false,
			wantDetected:    DetectionStateInstalled,
			wantLogContains: "successfully started",
		},
		{
			name: "NetcaExitNonzero_LeavesPartialAndErrors",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(lsnrctl).Returns(1, lsnrctlAbsent, "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(netca).Returns(2, "TNS-04404: response file not found\n", "")
			},
			spec:            validNetcaSpec(),
			wantErr:         "exit 2",
			wantDetected:    DetectionStatePartial,
			wantLogContains: "TNS-04404",
		},
		{
			name: "Reset_OnInstalled_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(0, "", "")
			},
			spec:         validNetcaSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "Reset_OnPartial_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(lsnrctl).Returns(1, lsnrctlAbsent, "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validNetcaSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStatePartial,
		},
		{
			name:      "RejectsMissingResponseFile",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: func() NetcaSpec {
				s := validNetcaSpec()
				s.ResponseFilePath = ""
				return s
			}(),
			wantErr:      "response_file_path required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingTarget",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "L", Port: 1521,
			},
			wantErr:      "target is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingOracleBase",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", ResponseFilePath: "/r"},
				ListenerName: "L", Port: 1521,
			},
			wantErr:      "oracle_base is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingListenerName",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				Port:        1521,
			},
			wantErr:      "listener_name is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsListenerNameWithShellMetachar_Semicolon",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "LISTENER;rm -rf /", Port: 1521,
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsListenerNameWithBacktick",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "L`whoami`", Port: 1521,
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsListenerNameWithDollar",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "L$(id)", Port: 1521,
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsListenerNameWithSpace",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "MY LISTENER", Port: 1521,
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsTargetWithNewline",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x\nrm -rf /", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "L", Port: 1521,
			},
			wantErr:      "control character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsPortZero",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "L", Port: 0,
			},
			wantErr:      "port must be",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsPortOutOfRange",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: NetcaSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				ListenerName: "L", Port: 99999,
			},
			wantErr:      "port must be",
			wantDetected: DetectionStateUnset,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := netcaSilentWithExec(context.Background(), mock, tc.spec, tc.reset)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				if tc.wantLogContains != "" && res != nil {
					assert.Contains(t, res.LogTail, tc.wantLogContains)
				}
				if tc.wantDetected != DetectionStateUnset && res != nil {
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

// netcaCtxCancelExec mimics asmca_test.go's helper: lets the first N probe
// calls go to the inner mock, then cancels ctx and fails the (N+1)-th
// call to simulate a transport disconnect mid-run.
type netcaCtxCancelExec struct {
	inner       host.Executor
	cancelAfter int
	calls       int
	cancel      context.CancelFunc
	wrapErr     error
}

func (e *netcaCtxCancelExec) Run(ctx context.Context, cmd string) (*host.RunResult, error) {
	e.calls++
	if e.calls <= e.cancelAfter {
		return e.inner.Run(ctx, cmd)
	}
	if e.cancel != nil {
		e.cancel()
	}
	return nil, e.wrapErr
}

func TestNetcaSilent_CtxCancelled_ReportsPartial(t *testing.T) {
	const partial = "/u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.partial"
	const installed = "/u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.installed"
	const lsnrctl = "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/lsnrctl status LISTENER"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/netca.LISTENER.partial"

	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f " + installed).Returns(1, "", "")
	mock.OnCommand(lsnrctl).Returns(1, "TNS-12541: TNS:no listener\n", "")
	mock.OnCommand("test -f " + partial).Returns(1, "", "")
	mock.OnCommand(mkdir).Returns(0, "", "")
	// netca call is intercepted and ctx-cancelled by the wrapper.

	ctx, cancel := context.WithCancel(context.Background())
	ex := &netcaCtxCancelExec{
		inner:       mock,
		cancelAfter: 4, // 3 probes (installed, lsnrctl, partial) + 1 sentinel-write
		cancel:      cancel,
		wrapErr:     errors.New("transport: connection closed"),
	}

	res, err := netcaSilentWithExec(ctx, ex, validNetcaSpec(), false)
	require.Error(t, err)
	require.NotNil(t, res)
	assert.Equal(t, DetectionStatePartial, res.Detected)
	assert.Contains(t, err.Error(), "interrupted")
	assert.Contains(t, err.Error(), "may still be running")
}

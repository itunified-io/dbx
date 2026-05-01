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

// validDbcaSpec returns a minimal valid spec usable across tests.
func validDbcaSpec() DbcaCreateDbSpec {
	return DbcaCreateDbSpec{
		InstallSpec: InstallSpec{
			Target:           "ext3adm1",
			OracleHome:       "/u01/app/oracle/product/19c/dbhome_1",
			OracleBase:       "/u01/app/oracle",
			ResponseFilePath: "/tmp/dbca.rsp",
		},
		DbUniqueName: "ORCL",
	}
}

func TestDbcaCreateDb(t *testing.T) {
	const partial = "/u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.partial"
	const installed = "/u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.installed"
	const srvctl = "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/srvctl status database -d ORCL"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.partial"
	const dbca = "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/dbca -silent -createDatabase -responseFile /tmp/dbca.rsp"
	const mv = "mv /u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.partial /u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.installed"

	cases := []struct {
		name            string
		setupMock       func(*hosttest.MockExecutor)
		spec            DbcaCreateDbSpec
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
			spec:         validDbcaSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "LiveDatabaseDetected_NoSentinel_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(srvctl).Returns(0, "Database is running.\n", "")
			},
			spec:         validDbcaSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "PartialSentinelDetected_ReturnsError",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(srvctl).Returns(1, "PRCD-1120: ...\n", "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validDbcaSpec(),
			wantErr:      "partial dbca state",
			wantDetected: DetectionStatePartial,
		},
		{
			name: "AbsentDatabase_RunsFullPipeline",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(srvctl).Returns(1, "PRCD-1120: ...\n", "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(dbca).Returns(0, "Prepare for db operation\n100% complete\nDatabase creation complete.\n", "")
				m.OnCommand(mv).Returns(0, "", "")
			},
			spec:            validDbcaSpec(),
			wantSkipped:     false,
			wantDetected:    DetectionStateInstalled,
			wantLogContains: "Database creation complete",
		},
		{
			name: "DbcaExitNonzero_LeavesPartialAndErrors",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(srvctl).Returns(1, "PRCD-1120: ...\n", "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(dbca).Returns(1, "[FATAL] DBT-06604: insufficient memory\n", "")
			},
			spec:            validDbcaSpec(),
			wantErr:         "exit 1",
			wantDetected:    DetectionStatePartial,
			wantLogContains: "DBT-06604",
		},
		{
			name: "Reset_OnInstalled_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(0, "", "")
			},
			spec:         validDbcaSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "Reset_OnPartial_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(srvctl).Returns(1, "PRCD-1120: ...\n", "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validDbcaSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStatePartial,
		},
		{
			name:      "RejectsMissingResponseFile",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: func() DbcaCreateDbSpec {
				s := validDbcaSpec()
				s.ResponseFilePath = ""
				return s
			}(),
			wantErr:      "response_file_path is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingTarget",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec:  InstallSpec{OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				DbUniqueName: "ORCL",
			},
			wantErr:      "target is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingOracleBase",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", ResponseFilePath: "/r"},
				DbUniqueName: "ORCL",
			},
			wantErr:      "oracle_base is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingDbUniqueName",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
			},
			wantErr:      "db_unique_name is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsDbUniqueNameWithShellMetachar_Semicolon",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				DbUniqueName: "ORCL;rm -rf /",
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsDbUniqueNameWithBacktick",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				DbUniqueName: "ORCL`whoami`",
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsDbUniqueNameWithDollar",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec:  InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				DbUniqueName: "ORCL$(id)",
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsTargetWithNewline",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec:  InstallSpec{Target: "x\nrm -rf /", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				DbUniqueName: "ORCL",
			},
			wantErr:      "control character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsSysPasswordFileWithNewline",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: DbcaCreateDbSpec{
				InstallSpec:     InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y", ResponseFilePath: "/r"},
				DbUniqueName:    "ORCL",
				SysPasswordFile: "/tmp/syspw\nrm -rf /",
			},
			wantErr:      "control character",
			wantDetected: DetectionStateUnset,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := dbcaCreateDbWithExec(context.Background(), mock, tc.spec, tc.reset)
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

// dbcaCtxCancelExec mimics netca_test.go's helper: lets the first N probe
// calls go to the inner mock, then cancels ctx and fails the (N+1)-th
// call to simulate a transport disconnect mid-run.
type dbcaCtxCancelExec struct {
	inner       host.Executor
	cancelAfter int
	calls       int
	cancel      context.CancelFunc
	wrapErr     error
}

func (e *dbcaCtxCancelExec) Run(ctx context.Context, cmd string) (*host.RunResult, error) {
	e.calls++
	if e.calls <= e.cancelAfter {
		return e.inner.Run(ctx, cmd)
	}
	if e.cancel != nil {
		e.cancel()
	}
	return nil, e.wrapErr
}

func TestDbcaCreateDb_CtxCancelled_ReportsPartial(t *testing.T) {
	const partial = "/u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.partial"
	const installed = "/u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.installed"
	const srvctl = "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/srvctl status database -d ORCL"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/dbca.ORCL.partial"

	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f " + installed).Returns(1, "", "")
	mock.OnCommand(srvctl).Returns(1, "PRCD-1120: ...\n", "")
	mock.OnCommand("test -f " + partial).Returns(1, "", "")
	mock.OnCommand(mkdir).Returns(0, "", "")
	// dbca call is intercepted and ctx-cancelled by the wrapper.

	ctx, cancel := context.WithCancel(context.Background())
	ex := &dbcaCtxCancelExec{
		inner:       mock,
		cancelAfter: 4, // 3 probes (installed, srvctl, partial) + 1 sentinel-write
		cancel:      cancel,
		wrapErr:     errors.New("transport: connection closed"),
	}

	res, err := dbcaCreateDbWithExec(ctx, ex, validDbcaSpec(), false)
	require.Error(t, err)
	require.NotNil(t, res)
	assert.Equal(t, DetectionStatePartial, res.Detected)
	assert.Contains(t, err.Error(), "interrupted")
	assert.Contains(t, err.Error(), "may still be running")
}

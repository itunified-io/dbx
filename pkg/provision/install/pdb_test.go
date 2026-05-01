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

// validPdbSpec returns a minimal valid spec usable across tests.
func validPdbSpec() PdbCreateSpec {
	return PdbCreateSpec{
		InstallSpec: InstallSpec{
			Target:     "ext3adm1",
			OracleHome: "/u01/app/oracle/product/19c/dbhome_1",
			OracleBase: "/u01/app/oracle",
		},
		CdbName:           "ORCL",
		PdbName:           "PDB1",
		AdminPasswordFile: "/tmp/pdbadmin.pw",
	}
}

func TestPdbCreate(t *testing.T) {
	const partial = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"
	const installed = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.installed"
	// SQL probe is delivered via heredoc-style echo | sqlplus
	const probeSQL = "set heading off pagesize 0 feedback off; select name from v$pdbs where name = upper('PDB1'); exit;"
	probeCmd := "echo " + shellEscape(probeSQL) + " | env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/sqlplus -s / as sysdba"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"
	const dbca = "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/dbca -silent -createPluggableDatabase -sourceDB ORCL -pdbName PDB1 -pdbAdminUserName PDBADMIN -pdbAdminPasswordFile /tmp/pdbadmin.pw"
	const mv = "mv /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.installed"

	cases := []struct {
		name            string
		setupMock       func(*hosttest.MockExecutor)
		spec            PdbCreateSpec
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
			spec:         validPdbSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "LivePdbDetected_NoSentinel_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(probeCmd).Returns(0, "PDB1\n", "")
			},
			spec:         validPdbSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "PartialSentinelDetected_ReturnsError",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(probeCmd).Returns(0, "\n", "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validPdbSpec(),
			wantErr:      "partial pdb state",
			wantDetected: DetectionStatePartial,
		},
		{
			name: "AbsentPdb_RunsFullPipeline",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(probeCmd).Returns(0, "\n", "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(dbca).Returns(0, "Prepare for db operation\n100% complete\nPluggable database creation complete.\n", "")
				m.OnCommand(mv).Returns(0, "", "")
			},
			spec:            validPdbSpec(),
			wantSkipped:     false,
			wantDetected:    DetectionStateInstalled,
			wantLogContains: "Pluggable database creation complete",
		},
		{
			name: "DbcaExitNonzero_LeavesPartialAndErrors",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(probeCmd).Returns(0, "\n", "")
				m.OnCommand("test -f " + partial).Returns(1, "", "")
				m.OnCommand(mkdir).Returns(0, "", "")
				m.OnCommand(dbca).Returns(1, "[FATAL] DBT-50000: insufficient resources\n", "")
			},
			spec:            validPdbSpec(),
			wantErr:         "exit 1",
			wantDetected:    DetectionStatePartial,
			wantLogContains: "DBT-50000",
		},
		{
			name: "Reset_OnInstalled_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(0, "", "")
			},
			spec:         validPdbSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "Reset_OnPartial_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installed).Returns(1, "", "")
				m.OnCommand(probeCmd).Returns(0, "\n", "")
				m.OnCommand("test -f " + partial).Returns(0, "", "")
			},
			spec:         validPdbSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStatePartial,
		},
		{
			name:      "RejectsMissingTarget",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{OracleHome: "/x", OracleBase: "/y"},
				CdbName:           "ORCL",
				PdbName:           "PDB1",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "target is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingOracleBase",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x", OracleHome: "/x"},
				CdbName:           "ORCL",
				PdbName:           "PDB1",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "oracle_base is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingCdbName",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				PdbName:           "PDB1",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "cdb_name is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingPdbName",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				CdbName:           "ORCL",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "pdb_name is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsMissingAdminPasswordFile",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec: InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				CdbName:     "ORCL",
				PdbName:     "PDB1",
			},
			wantErr:      "admin_password_file is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsCdbNameWithShellMetachar_Semicolon",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				CdbName:           "ORCL;rm -rf /",
				PdbName:           "PDB1",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsPdbNameWithBacktick",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				CdbName:           "ORCL",
				PdbName:           "PDB1`whoami`",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsPdbNameWithDollar",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				CdbName:           "ORCL",
				PdbName:           "PDB1$(id)",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsTargetWithNewline",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x\nrm -rf /", OracleHome: "/x", OracleBase: "/y"},
				CdbName:           "ORCL",
				PdbName:           "PDB1",
				AdminPasswordFile: "/tmp/pw",
			},
			wantErr:      "control character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsAdminPasswordFileWithNewline",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: PdbCreateSpec{
				InstallSpec:       InstallSpec{Target: "x", OracleHome: "/x", OracleBase: "/y"},
				CdbName:           "ORCL",
				PdbName:           "PDB1",
				AdminPasswordFile: "/tmp/pw\nrm -rf /",
			},
			wantErr:      "control character",
			wantDetected: DetectionStateUnset,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := pdbCreateWithExec(context.Background(), mock, tc.spec, tc.reset)
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

// TestPdbCreate_WithResponseFile asserts the alternate dbca CLI shape
// (-responseFile) is selected when ResponseFilePath is set.
func TestPdbCreate_WithResponseFile(t *testing.T) {
	spec := validPdbSpec()
	spec.ResponseFilePath = "/tmp/pdb.rsp"

	const partial = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"
	const installed = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.installed"
	const probeSQL = "set heading off pagesize 0 feedback off; select name from v$pdbs where name = upper('PDB1'); exit;"
	probeCmd := "echo " + shellEscape(probeSQL) + " | env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/sqlplus -s / as sysdba"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"
	const dbca = "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/dbca -silent -createPluggableDatabase -responseFile /tmp/pdb.rsp"
	const mv = "mv /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.installed"

	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f " + installed).Returns(1, "", "")
	mock.OnCommand(probeCmd).Returns(0, "\n", "")
	mock.OnCommand("test -f " + partial).Returns(1, "", "")
	mock.OnCommand(mkdir).Returns(0, "", "")
	mock.OnCommand(dbca).Returns(0, "Pluggable database creation complete.\n", "")
	mock.OnCommand(mv).Returns(0, "", "")

	res, err := pdbCreateWithExec(context.Background(), mock, spec, false)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Skipped)
	assert.Equal(t, DetectionStateInstalled, res.Detected)
}

// TestPdbCreate_WithDatafileDest asserts -pdbDatafileDestination is
// appended when DatafileDest is set.
func TestPdbCreate_WithDatafileDest(t *testing.T) {
	spec := validPdbSpec()
	spec.DatafileDest = "+DATA"

	const installed = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.installed"
	const partial = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"
	const probeSQL = "set heading off pagesize 0 feedback off; select name from v$pdbs where name = upper('PDB1'); exit;"
	probeCmd := "echo " + shellEscape(probeSQL) + " | env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/sqlplus -s / as sysdba"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"
	dbca := "env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/dbca -silent -createPluggableDatabase -sourceDB ORCL -pdbName PDB1 -pdbAdminUserName PDBADMIN -pdbAdminPasswordFile /tmp/pdbadmin.pw -pdbDatafileDestination " + shellEscape("+DATA")
	const mv = "mv /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.installed"

	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f " + installed).Returns(1, "", "")
	mock.OnCommand(probeCmd).Returns(0, "\n", "")
	mock.OnCommand("test -f " + partial).Returns(1, "", "")
	mock.OnCommand(mkdir).Returns(0, "", "")
	mock.OnCommand(dbca).Returns(0, "OK\n", "")
	mock.OnCommand(mv).Returns(0, "", "")

	res, err := pdbCreateWithExec(context.Background(), mock, spec, false)
	require.NoError(t, err)
	assert.Equal(t, DetectionStateInstalled, res.Detected)
}

// pdbCtxCancelExec — same pattern as dbcaCtxCancelExec.
type pdbCtxCancelExec struct {
	inner       host.Executor
	cancelAfter int
	calls       int
	cancel      context.CancelFunc
	wrapErr     error
}

func (e *pdbCtxCancelExec) Run(ctx context.Context, cmd string) (*host.RunResult, error) {
	e.calls++
	if e.calls <= e.cancelAfter {
		return e.inner.Run(ctx, cmd)
	}
	if e.cancel != nil {
		e.cancel()
	}
	return nil, e.wrapErr
}

func TestPdbCreate_CtxCancelled_ReportsPartial(t *testing.T) {
	const installed = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.installed"
	const partial = "/u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"
	const probeSQL = "set heading off pagesize 0 feedback off; select name from v$pdbs where name = upper('PDB1'); exit;"
	probeCmd := "echo " + shellEscape(probeSQL) + " | env ORACLE_HOME=/u01/app/oracle/product/19c/dbhome_1 /u01/app/oracle/product/19c/dbhome_1/bin/sqlplus -s / as sysdba"
	const mkdir = "mkdir -p /u01/app/oracle/cfgtoollogs/dbx && : > /u01/app/oracle/cfgtoollogs/dbx/pdb.ORCL.PDB1.partial"

	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f " + installed).Returns(1, "", "")
	mock.OnCommand(probeCmd).Returns(0, "\n", "")
	mock.OnCommand("test -f " + partial).Returns(1, "", "")
	mock.OnCommand(mkdir).Returns(0, "", "")

	ctx, cancel := context.WithCancel(context.Background())
	ex := &pdbCtxCancelExec{
		inner:       mock,
		cancelAfter: 4, // 3 probes (installed, sqlplus, partial) + 1 sentinel-write
		cancel:      cancel,
		wrapErr:     errors.New("transport: connection closed"),
	}

	res, err := pdbCreateWithExec(ctx, ex, validPdbSpec(), false)
	require.Error(t, err)
	require.NotNil(t, res)
	assert.Equal(t, DetectionStatePartial, res.Detected)
	assert.Contains(t, err.Error(), "interrupted")
	assert.Contains(t, err.Error(), "may still be running")
}

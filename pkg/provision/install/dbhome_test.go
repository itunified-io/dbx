package install

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/host/hosttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBHomeInstall(t *testing.T) {
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
			name: "DetectsExistingInstall_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/bin/oracle -V 2>&1 | head -1").
					Returns(0, "Oracle Database 19.0.0.0.0 - Production\nVersion 19.26.0.0.0\n", "")
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/OPatch/opatch lsinventory 2>&1 | head -5").
					Returns(0, "Oracle Interim Patch Installer version 12.2.0.1.46\n", "")
			},
			spec:         InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/oracle/product/19c/dbhome_1"},
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "AbsentState_RunsInstaller",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/bin/oracle -V 2>&1 | head -1").Returns(127, "", "No such file")
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/OPatch/opatch lsinventory 2>&1 | head -5").Returns(127, "", "")
				m.OnCommandPattern(`/smb/software/oracle/19c/db_home/runInstaller -silent -responseFile /tmp/dbhome\.rsp.*`).
					Returns(0, "Successfully Setup Software.\n", "")
			},
			spec: InstallSpec{
				Target:           "ext3adm1",
				OracleHome:       "/u01/app/oracle/product/19c/dbhome_1",
				OracleBase:       "/u01/app/oracle",
				SoftwareStaging:  "/smb/software/oracle/19c/db_home",
				ResponseFilePath: "/tmp/dbhome.rsp",
			},
			wantSkipped:     false,
			wantDetected:    DetectionStateAbsent,
			wantLogContains: "Successfully Setup Software",
		},
		{
			name:      "RejectsInvalidSpec",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec:      InstallSpec{},
			wantErr:   "target is required",
		},
		{
			name: "PartialState_AbortsWithoutReset",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/bin/oracle -V 2>&1 | head -1").
					Returns(0, "Oracle Database 19.0.0.0.0\n", "")
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/OPatch/opatch lsinventory 2>&1 | head -5").Returns(127, "", "")
			},
			spec:    InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/oracle/product/19c/dbhome_1"},
			wantErr: "partial dbhome install detected",
		},
		{
			name: "DetectsExistingInstall_Skips_23ai",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("/u01/app/oracle/product/23ai/dbhome_1/bin/oracle -V 2>&1 | head -1").
					Returns(0, "Oracle Database 23.0.0.0.0 - Production\nVersion 23.5.0.0.0\n", "")
				m.OnCommand("/u01/app/oracle/product/23ai/dbhome_1/OPatch/opatch lsinventory 2>&1 | head -5").
					Returns(0, "Oracle Interim Patch Installer version 12.2.0.1.46\n", "")
			},
			spec:         InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/oracle/product/23ai/dbhome_1"},
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "RunInstallerNonzero_ReturnsError",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/bin/oracle -V 2>&1 | head -1").Returns(127, "", "")
				m.OnCommand("/u01/app/oracle/product/19c/dbhome_1/OPatch/opatch lsinventory 2>&1 | head -5").Returns(127, "", "")
				m.OnCommandPattern(`.*runInstaller.*`).Returns(254, "INFO: Skipping the prereq checks.\nERROR: Some files failed to copy.\n", "")
			},
			spec: InstallSpec{
				Target:           "ext3adm1",
				OracleHome:       "/u01/app/oracle/product/19c/dbhome_1",
				SoftwareStaging:  "/smb/software/oracle/19c/db_home",
				ResponseFilePath: "/tmp/dbhome.rsp",
			},
			wantErr:         "exit 254",
			wantLogContains: "Some files failed to copy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := dbhomeInstallWithExec(context.Background(), mock, tc.spec, tc.reset)
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

package install

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/host/hosttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGridInstall(t *testing.T) {
	cases := []struct {
		name            string
		setupMock       func(*hosttest.MockExecutor)
		spec            InstallSpec
		reset           bool
		wantErr         string // empty = no error
		wantSkipped     bool
		wantDetected    DetectionState
		wantLogContains string
	}{
		{
			name: "DetectsExistingInstall_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f /etc/oraInst.loc && cat /etc/oraInst.loc").
					Returns(0, "inventory_loc=/u01/app/oraInventory\ninst_group=oinstall\n", "")
				m.OnCommand("test -d /u01/app/19c/grid/inventory && ls -A /u01/app/19c/grid/inventory | head -1").
					Returns(0, "ContentsXML\n", "")
			},
			spec:         InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid", OracleBase: "/u01/app/grid"},
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "AbsentState_RunsInstaller",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f /etc/oraInst.loc && cat /etc/oraInst.loc").Returns(1, "", "")
				m.OnCommand("test -d /u01/app/19c/grid/inventory && ls -A /u01/app/19c/grid/inventory | head -1").
					Returns(1, "", "")
				m.OnCommandPattern(`/u01/app/19c/grid/runInstaller -silent -responseFile /tmp/grid\.rsp.*`).
					Returns(0, "Installation Successful.\n", "")
			},
			spec: InstallSpec{
				Target: "ext3adm1", OracleHome: "/u01/app/19c/grid",
				OracleBase: "/u01/app/grid", ResponseFilePath: "/tmp/grid.rsp",
			},
			wantSkipped:     false,
			wantDetected:    DetectionStateAbsent,
			wantLogContains: "Installation Successful",
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
				m.OnCommand("test -f /etc/oraInst.loc && cat /etc/oraInst.loc").
					Returns(0, "inventory_loc=/u01/app/oraInventory\n", "")
				m.OnCommand("test -d /u01/app/19c/grid/inventory && ls -A /u01/app/19c/grid/inventory | head -1").
					Returns(1, "", "")
			},
			spec:    InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"},
			wantErr: "partial install detected",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := gridInstallWithExec(context.Background(), mock, tc.spec, tc.reset)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
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

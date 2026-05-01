package install

import (
	"context"
	"testing"

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
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile && cat /u01/app/19c/grid/install/rootsh.touchfile").
					Returns(0, "ran 2026-04-30T10:00:00Z\n", "")
			},
			spec:         InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"},
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "AbsentTouchfile_RunsScript",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile && cat /u01/app/19c/grid/install/rootsh.touchfile").
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
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile && cat /u01/app/19c/grid/install/rootsh.touchfile").
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
				m.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile && cat /u01/app/19c/grid/install/rootsh.touchfile").
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

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

// validAsmlibSpec returns a minimal valid ASMlib spec usable across tests.
func validAsmlibSpec() AsmDiskLabelSpec {
	return AsmDiskLabelSpec{
		Target:         "ext3adm1",
		GridHome:       "/u01/app/19c/grid",
		OracleBase:     "/u01/app/grid",
		Implementation: AsmDiskLabelImplAsmlib,
		Labels: []AsmLabelEntry{
			{Name: "DATA1", Device: "/dev/sdb"},
		},
	}
}

// validAFDSpec returns a minimal valid AFD spec usable across tests.
func validAFDSpec() AsmDiskLabelSpec {
	return AsmDiskLabelSpec{
		Target:         "ext3adm1",
		GridHome:       "/u01/app/19c/grid",
		OracleBase:     "/u01/app/grid",
		Implementation: AsmDiskLabelImplAFD,
		Labels: []AsmLabelEntry{
			{Name: "DATA1", Device: "/dev/sdb"},
		},
	}
}

func TestAsmDiskLabel(t *testing.T) {
	const installedAsmlib = "/u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.installed"
	const partialAsmlib = "/u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.partial"
	const probeAsmlib = "/usr/sbin/oracleasm listdisks"
	const mkdirAsmlib = "mkdir -p /u01/app/grid/cfgtoollogs/dbx && : > /u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.partial"
	const createAsmlib = "/usr/sbin/oracleasm createdisk DATA1 /dev/sdb"
	const mvAsmlib = "mv /u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.partial /u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.installed"

	const probeAFD = "env ORACLE_HOME=/u01/app/19c/grid /u01/app/19c/grid/bin/asmcmd afd_lslbl /dev/sdb"
	const createAFD = "env ORACLE_HOME=/u01/app/19c/grid /u01/app/19c/grid/bin/asmcmd afd_label DATA1 /dev/sdb --init"

	cases := []struct {
		name            string
		setupMock       func(*hosttest.MockExecutor)
		spec            AsmDiskLabelSpec
		reset           bool
		wantErr         string
		wantSkipped     bool   // for first label
		wantDetected    DetectionState // for first label
		wantLogContains string
	}{
		{
			name: "Asmlib_InstalledSentinel_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(0, "", "")
			},
			spec:         validAsmlibSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "Asmlib_LiveProbeMatches_NoSentinel_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(1, "", "")
				m.OnCommand(probeAsmlib).Returns(0, "DATA1\nDATA2\n", "")
			},
			spec:         validAsmlibSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "Asmlib_PartialSentinel_ReturnsError",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(1, "", "")
				m.OnCommand(probeAsmlib).Returns(0, "OTHER\n", "")
				m.OnCommand("test -f " + partialAsmlib).Returns(0, "", "")
			},
			spec:         validAsmlibSpec(),
			wantErr:      "partial asm-label state",
			wantDetected: DetectionStatePartial,
		},
		{
			name: "Asmlib_Absent_RunsFullPipeline",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(1, "", "")
				m.OnCommand(probeAsmlib).Returns(0, "OTHER\n", "")
				m.OnCommand("test -f " + partialAsmlib).Returns(1, "", "")
				m.OnCommand(mkdirAsmlib).Returns(0, "", "")
				m.OnCommand(createAsmlib).Returns(0, "Writing disk header: done\nInstantiating disk: done\n", "")
				m.OnCommand(mvAsmlib).Returns(0, "", "")
			},
			spec:            validAsmlibSpec(),
			wantSkipped:     false,
			wantDetected:    DetectionStateInstalled,
			wantLogContains: "Writing disk header",
		},
		{
			name: "Asmlib_CreateExitNonzero_LeavesPartialAndErrors",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(1, "", "")
				m.OnCommand(probeAsmlib).Returns(0, "OTHER\n", "")
				m.OnCommand("test -f " + partialAsmlib).Returns(1, "", "")
				m.OnCommand(mkdirAsmlib).Returns(0, "", "")
				m.OnCommand(createAsmlib).Returns(1, "ERROR: device not found\n", "")
			},
			spec:            validAsmlibSpec(),
			wantErr:         "exit 1",
			wantDetected:    DetectionStatePartial,
			wantLogContains: "device not found",
		},
		{
			name: "Asmlib_Reset_OnInstalled_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(0, "", "")
			},
			spec:         validAsmlibSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "Asmlib_Reset_OnPartial_PrintsRunbook_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(1, "", "")
				m.OnCommand(probeAsmlib).Returns(0, "OTHER\n", "")
				m.OnCommand("test -f " + partialAsmlib).Returns(0, "", "")
			},
			spec:         validAsmlibSpec(),
			reset:        true,
			wantSkipped:  true,
			wantDetected: DetectionStatePartial,
		},
		{
			name: "AFD_InstalledSentinel_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(0, "", "")
			},
			spec:         validAFDSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "AFD_LiveProbeMatches_NoSentinel_Skips",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(1, "", "")
				m.OnCommand(probeAFD).Returns(0, "--------------------------------------------------------------------------------\nLabel                     Duplicate  Path\n================================================================================\nDATA1                                /dev/sdb\n", "")
			},
			spec:         validAFDSpec(),
			wantSkipped:  true,
			wantDetected: DetectionStateInstalled,
		},
		{
			name: "AFD_Absent_RunsFullPipeline",
			setupMock: func(m *hosttest.MockExecutor) {
				m.OnCommand("test -f " + installedAsmlib).Returns(1, "", "")
				m.OnCommand(probeAFD).Returns(0, "Label                     Duplicate  Path\n", "")
				m.OnCommand("test -f " + partialAsmlib).Returns(1, "", "")
				m.OnCommand(mkdirAsmlib).Returns(0, "", "")
				m.OnCommand(createAFD).Returns(0, "Disk DATA1 successfully labeled.\n", "")
				m.OnCommand(mvAsmlib).Returns(0, "", "")
			},
			spec:            validAFDSpec(),
			wantSkipped:     false,
			wantDetected:    DetectionStateInstalled,
			wantLogContains: "successfully labeled",
		},
		{
			name:         "RejectsMissingTarget",
			setupMock:    func(m *hosttest.MockExecutor) {},
			spec:         AsmDiskLabelSpec{GridHome: "/x", OracleBase: "/y", Implementation: AsmDiskLabelImplAsmlib, Labels: []AsmLabelEntry{{Name: "X", Device: "/dev/sdb"}}},
			wantErr:      "target is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:         "RejectsMissingGridHome",
			setupMock:    func(m *hosttest.MockExecutor) {},
			spec:         AsmDiskLabelSpec{Target: "x", OracleBase: "/y", Implementation: AsmDiskLabelImplAsmlib, Labels: []AsmLabelEntry{{Name: "X", Device: "/dev/sdb"}}},
			wantErr:      "grid_home is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:         "RejectsMissingOracleBase",
			setupMock:    func(m *hosttest.MockExecutor) {},
			spec:         AsmDiskLabelSpec{Target: "x", GridHome: "/y", Implementation: AsmDiskLabelImplAsmlib, Labels: []AsmLabelEntry{{Name: "X", Device: "/dev/sdb"}}},
			wantErr:      "oracle_base is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:         "RejectsInvalidImpl",
			setupMock:    func(m *hosttest.MockExecutor) {},
			spec:         AsmDiskLabelSpec{Target: "x", GridHome: "/y", OracleBase: "/z", Implementation: "udev", Labels: []AsmLabelEntry{{Name: "X", Device: "/dev/sdb"}}},
			wantErr:      "implementation must be",
			wantDetected: DetectionStateUnset,
		},
		{
			name:         "RejectsEmptyLabels",
			setupMock:    func(m *hosttest.MockExecutor) {},
			spec:         AsmDiskLabelSpec{Target: "x", GridHome: "/y", OracleBase: "/z", Implementation: AsmDiskLabelImplAsmlib},
			wantErr:      "labels list is required",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsLabelNameWithSemicolon",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmDiskLabelSpec{
				Target: "x", GridHome: "/y", OracleBase: "/z", Implementation: AsmDiskLabelImplAsmlib,
				Labels: []AsmLabelEntry{{Name: "DATA1;rm -rf /", Device: "/dev/sdb"}},
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsLabelNameWithBacktick",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmDiskLabelSpec{
				Target: "x", GridHome: "/y", OracleBase: "/z", Implementation: AsmDiskLabelImplAsmlib,
				Labels: []AsmLabelEntry{{Name: "D`whoami`", Device: "/dev/sdb"}},
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsDeviceWithDollar",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmDiskLabelSpec{
				Target: "x", GridHome: "/y", OracleBase: "/z", Implementation: AsmDiskLabelImplAsmlib,
				Labels: []AsmLabelEntry{{Name: "DATA1", Device: "/dev/$(id)"}},
			},
			wantErr:      "disallowed character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsTargetWithNewline",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmDiskLabelSpec{
				Target: "x\nrm -rf /", GridHome: "/y", OracleBase: "/z", Implementation: AsmDiskLabelImplAsmlib,
				Labels: []AsmLabelEntry{{Name: "DATA1", Device: "/dev/sdb"}},
			},
			wantErr:      "control character",
			wantDetected: DetectionStateUnset,
		},
		{
			name:      "RejectsDuplicateLabelNames",
			setupMock: func(m *hosttest.MockExecutor) {},
			spec: AsmDiskLabelSpec{
				Target: "x", GridHome: "/y", OracleBase: "/z", Implementation: AsmDiskLabelImplAsmlib,
				Labels: []AsmLabelEntry{
					{Name: "DATA1", Device: "/dev/sdb"},
					{Name: "DATA1", Device: "/dev/sdc"},
				},
			},
			wantErr:      "duplicated",
			wantDetected: DetectionStateUnset,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := hosttest.NewMockExecutor()
			tc.setupMock(mock)
			res, err := asmDiskLabelWithExec(context.Background(), mock, tc.spec, tc.reset)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				if tc.wantLogContains != "" && res != nil && len(res.Labels) > 0 {
					assert.Contains(t, res.Labels[0].LogTail, tc.wantLogContains)
				}
				if tc.wantDetected != DetectionStateUnset && res != nil && len(res.Labels) > 0 {
					assert.Equal(t, tc.wantDetected, res.Labels[0].Detected)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Len(t, res.Labels, 1)
			assert.Equal(t, tc.wantSkipped, res.Labels[0].Skipped)
			assert.Equal(t, tc.wantDetected, res.Labels[0].Detected)
			if tc.wantLogContains != "" {
				assert.Contains(t, res.Labels[0].LogTail, tc.wantLogContains)
			}
		})
	}
}

// TestAsmDiskLabel_MultipleLabels_AllProcessed exercises the label loop
// to confirm a mix of skip + create works in a single invocation.
func TestAsmDiskLabel_MultipleLabels_AllProcessed(t *testing.T) {
	const installedData1 = "/u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.installed"
	const installedData2 = "/u01/app/grid/cfgtoollogs/dbx/asm-label.DATA2.installed"
	const partialData2 = "/u01/app/grid/cfgtoollogs/dbx/asm-label.DATA2.partial"
	const probe = "/usr/sbin/oracleasm listdisks"
	const mkdir2 = "mkdir -p /u01/app/grid/cfgtoollogs/dbx && : > /u01/app/grid/cfgtoollogs/dbx/asm-label.DATA2.partial"
	const create2 = "/usr/sbin/oracleasm createdisk DATA2 /dev/sdc"
	const mv2 = "mv /u01/app/grid/cfgtoollogs/dbx/asm-label.DATA2.partial /u01/app/grid/cfgtoollogs/dbx/asm-label.DATA2.installed"

	mock := hosttest.NewMockExecutor()
	// DATA1: installed sentinel present → skip.
	mock.OnCommand("test -f " + installedData1).Returns(0, "", "")
	// DATA2: absent → run full pipeline.
	mock.OnCommand("test -f " + installedData2).Returns(1, "", "")
	mock.OnCommand(probe).Returns(0, "DATA1\n", "")
	mock.OnCommand("test -f " + partialData2).Returns(1, "", "")
	mock.OnCommand(mkdir2).Returns(0, "", "")
	mock.OnCommand(create2).Returns(0, "ok\n", "")
	mock.OnCommand(mv2).Returns(0, "", "")

	spec := AsmDiskLabelSpec{
		Target:         "ext3adm1",
		GridHome:       "/u01/app/19c/grid",
		OracleBase:     "/u01/app/grid",
		Implementation: AsmDiskLabelImplAsmlib,
		Labels: []AsmLabelEntry{
			{Name: "DATA1", Device: "/dev/sdb"},
			{Name: "DATA2", Device: "/dev/sdc"},
		},
	}
	res, err := asmDiskLabelWithExec(context.Background(), mock, spec, false)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Labels, 2)
	assert.True(t, res.Labels[0].Skipped)
	assert.Equal(t, DetectionStateInstalled, res.Labels[0].Detected)
	assert.False(t, res.Labels[1].Skipped)
	assert.Equal(t, DetectionStateInstalled, res.Labels[1].Detected)
}

// asmLabelCtxCancelExec mimics netcaCtxCancelExec: lets the first N
// probe calls go through, then cancels ctx and fails the (N+1)-th call.
type asmLabelCtxCancelExec struct {
	inner       host.Executor
	cancelAfter int
	calls       int
	cancel      context.CancelFunc
	wrapErr     error
}

func (e *asmLabelCtxCancelExec) Run(ctx context.Context, cmd string) (*host.RunResult, error) {
	e.calls++
	if e.calls <= e.cancelAfter {
		return e.inner.Run(ctx, cmd)
	}
	if e.cancel != nil {
		e.cancel()
	}
	return nil, e.wrapErr
}

func TestAsmDiskLabel_CtxCancelled_ReportsPartial(t *testing.T) {
	const installed = "/u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.installed"
	const partial = "/u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.partial"
	const probe = "/usr/sbin/oracleasm listdisks"
	const mkdir = "mkdir -p /u01/app/grid/cfgtoollogs/dbx && : > /u01/app/grid/cfgtoollogs/dbx/asm-label.DATA1.partial"

	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f " + installed).Returns(1, "", "")
	mock.OnCommand(probe).Returns(0, "OTHER\n", "")
	mock.OnCommand("test -f " + partial).Returns(1, "", "")
	mock.OnCommand(mkdir).Returns(0, "", "")
	// createdisk call is intercepted and ctx-cancelled by the wrapper.

	ctx, cancel := context.WithCancel(context.Background())
	ex := &asmLabelCtxCancelExec{
		inner:       mock,
		cancelAfter: 4, // 3 probes + 1 sentinel-write
		cancel:      cancel,
		wrapErr:     errors.New("transport: connection closed"),
	}

	res, err := asmDiskLabelWithExec(ctx, ex, validAsmlibSpec(), false)
	require.Error(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Labels, 1)
	assert.Equal(t, DetectionStatePartial, res.Labels[0].Detected)
	assert.Contains(t, err.Error(), "interrupted")
	assert.Contains(t, err.Error(), "may still be running")
}

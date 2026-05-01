package install

import (
	"context"
	"sync"
	"testing"

	"github.com/itunified-io/dbx/pkg/host/hosttest"
	"github.com/itunified-io/dbx/pkg/otel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureExporter is a tiny in-test Exporter that retains every
// exported span for assertion. Concurrency-safe so deferred Export()
// from primitives racing with the test is fine.
type captureExporter struct {
	mu    sync.Mutex
	spans []otel.Span
}

func (e *captureExporter) Export(_ context.Context, spans []otel.Span) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spans = append(e.spans, spans...)
	return nil
}

func (e *captureExporter) Shutdown(_ context.Context) error { return nil }

func (e *captureExporter) Snapshot() []otel.Span {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]otel.Span, len(e.spans))
	copy(out, e.spans)
	return out
}

// withCaptureExporter installs a fresh capture exporter for the duration
// of a test, restoring the prior exporter on cleanup.
func withCaptureExporter(t *testing.T) *captureExporter {
	t.Helper()
	prev := otel.GlobalExporter()
	cap := &captureExporter{}
	otel.SetGlobalExporter(cap)
	t.Cleanup(func() { otel.SetGlobalExporter(prev) })
	return cap
}

// TestInstallPrimitives_EmitOTELSpan verifies each of the 8 install
// primitives emits exactly one OTEL span with the expected name +
// status + dbx.* attributes when invoked through the *WithExec test
// seam. Happy-path table covers the StatusOK case; the
// "MalformedSpec_StatusError" subtest covers StatusError.
func TestInstallPrimitives_EmitOTELSpan(t *testing.T) {
	t.Run("Grid_HappyPath", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		mock.OnCommand("test -f /etc/oraInst.loc").Returns(0, "x\n", "")
		mock.OnCommand("test -d /u01/app/19c/grid/inventory && ls -A /u01/app/19c/grid/inventory | head -1").
			Returns(0, "ContentsXML\n", "")
		_, err := gridInstallWithExec(context.Background(), mock,
			InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid", OracleBase: "/u01/app/grid"},
			false)
		require.NoError(t, err)
		assertOneSpan(t, cap, "provision.install.grid", otel.StatusOK, map[string]string{
			otel.AttrDbxHost:       "ext3adm1",
			otel.AttrDbxEntityType: "oracle_grid_home",
			otel.AttrDbxEntityName: "/u01/app/19c/grid",
		})
	})

	t.Run("Grid_ValidateError_StatusError", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		_, err := gridInstallWithExec(context.Background(), mock, InstallSpec{}, false)
		require.Error(t, err)
		spans := cap.Snapshot()
		require.Len(t, spans, 1)
		assert.Equal(t, "provision.install.grid", spans[0].Name)
		assert.Equal(t, otel.StatusError, spans[0].Status)
	})

	t.Run("DBHome_HappyPath", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		mock.OnCommand("/u01/app/19c/dbhome_1/bin/oracle -V 2>&1 | head -1").Returns(0, "Oracle Database 19c\n", "")
		mock.OnCommand("/u01/app/19c/dbhome_1/OPatch/opatch lsinventory 2>&1 | head -5").Returns(0, "Installed Top-level Products\n", "")
		_, err := dbhomeInstallWithExec(context.Background(), mock,
			InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/dbhome_1", OracleBase: "/u01/app/oracle"},
			false)
		require.NoError(t, err)
		assertOneSpan(t, cap, "provision.install.dbhome", otel.StatusOK, map[string]string{
			otel.AttrDbxEntityType: "oracle_db_home",
		})
	})

	t.Run("RootSh_HappyPath_AlreadyDone", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		mock.OnCommand("test -f /u01/app/19c/grid/install/rootsh.touchfile").Returns(0, "", "")
		_, err := rootShWithExec(context.Background(), mock,
			InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"},
			false)
		require.NoError(t, err)
		assertOneSpan(t, cap, "provision.install.root_sh", otel.StatusOK, map[string]string{
			otel.AttrDbxEntityType: "oracle_root_sh",
		})
	})

	t.Run("Asmca_ValidateError_StatusError", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		_, err := asmcaSilentWithExec(context.Background(), mock, AsmcaSpec{}, false)
		require.Error(t, err)
		spans := cap.Snapshot()
		require.Len(t, spans, 1)
		assert.Equal(t, "provision.install.asmca", spans[0].Name)
		assert.Equal(t, otel.StatusError, spans[0].Status)
	})

	t.Run("Netca_ValidateError_StatusError", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		_, err := netcaSilentWithExec(context.Background(), mock, NetcaSpec{}, false)
		require.Error(t, err)
		spans := cap.Snapshot()
		require.Len(t, spans, 1)
		assert.Equal(t, "provision.install.netca", spans[0].Name)
		assert.Equal(t, otel.StatusError, spans[0].Status)
	})

	t.Run("AsmLabel_ValidateError_StatusError", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		_, err := asmDiskLabelWithExec(context.Background(), mock, AsmDiskLabelSpec{}, false)
		require.Error(t, err)
		spans := cap.Snapshot()
		require.Len(t, spans, 1)
		assert.Equal(t, "provision.install.asm_label", spans[0].Name)
		assert.Equal(t, otel.StatusError, spans[0].Status)
	})

	t.Run("Dbca_ValidateError_StatusError", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		_, err := dbcaCreateDbWithExec(context.Background(), mock, DbcaCreateDbSpec{}, false)
		require.Error(t, err)
		spans := cap.Snapshot()
		require.Len(t, spans, 1)
		assert.Equal(t, "provision.install.dbca_create_db", spans[0].Name)
		assert.Equal(t, otel.StatusError, spans[0].Status)
	})

	t.Run("Pdb_ValidateError_StatusError", func(t *testing.T) {
		cap := withCaptureExporter(t)
		mock := hosttest.NewMockExecutor()
		_, err := pdbCreateWithExec(context.Background(), mock, PdbCreateSpec{}, false)
		require.Error(t, err)
		spans := cap.Snapshot()
		require.Len(t, spans, 1)
		assert.Equal(t, "provision.install.pdb_create", spans[0].Name)
		assert.Equal(t, otel.StatusError, spans[0].Status)
	})
}

// assertOneSpan asserts the capture exporter saw exactly one span with
// the given name and status, and that every (key, value) in wantAttrs
// is present on the span's attribute map.
func assertOneSpan(t *testing.T, cap *captureExporter, wantName string, wantStatus otel.Status, wantAttrs map[string]string) {
	t.Helper()
	spans := cap.Snapshot()
	require.Len(t, spans, 1, "expected exactly one span")
	assert.Equal(t, wantName, spans[0].Name)
	assert.Equal(t, wantStatus, spans[0].Status)
	got := spans[0].AttributeMap()
	for k, v := range wantAttrs {
		assert.Equal(t, v, got[k], "attribute %s mismatch", k)
	}
}

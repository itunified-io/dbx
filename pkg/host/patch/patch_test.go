package patch_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/patch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePatchStatus(t *testing.T) {
	updates := `FEDORA-2026-abc123 Important/Sec. kernel-5.15.0-201.el8.x86_64
FEDORA-2026-def456 Critical/Sec.  openssl-1.1.1k-13.el8_9.x86_64
`
	status, err := patch.ParsePatchStatus(updates)
	require.NoError(t, err)
	assert.Equal(t, 2, status.TotalUpdates)
	assert.Equal(t, 2, status.SecurityUpdates)
	assert.Equal(t, 1, status.CriticalCount)
	assert.Equal(t, 1, status.ImportantCount)
}

func TestParseKspliceStatus(t *testing.T) {
	output := `Effective kernel version: 5.15.0-200.el8.x86_64
Installed updates:
  Known-exploit detection
  CVE-2026-1234 [Important]
  CVE-2026-5678 [Critical]
`
	ks, err := patch.ParseKspliceStatus(output)
	require.NoError(t, err)
	assert.Equal(t, "5.15.0-200.el8.x86_64", ks.EffectiveKernel)
	assert.Len(t, ks.InstalledPatches, 2)
	assert.Equal(t, "CVE-2026-1234", ks.InstalledPatches[0].CVE)
	assert.Equal(t, "Important", ks.InstalledPatches[0].Severity)
}

func TestParseLastUpdateTime(t *testing.T) {
	output := "openssl-1.1.1k-12.el8_9.x86_64  Wed 09 Apr 2026 02:00:00 PM UTC\n"
	ts, err := patch.ParseLastUpdateTime(output)
	require.NoError(t, err)
	assert.Equal(t, 2026, ts.Year())
	assert.Equal(t, 4, int(ts.Month()))
}

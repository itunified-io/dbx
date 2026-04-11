package metrics_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProcStat(t *testing.T) {
	content := `cpu  74135 1022 18563 1121456 2140 6230 3104 0 0 0
cpu0 18200 260 4620 280400 530 1560 780 0 0 0
cpu1 18900 255 4650 279800 540 1570 790 0 0 0
cpu2 18500 250 4640 280100 535 1550 770 0 0 0
cpu3 18535 257 4653 281156 535 1550 764 0 0 0
`
	m, err := metrics.ParseProcStat(content)
	require.NoError(t, err)
	assert.Equal(t, uint64(74135), m.User)
	assert.Equal(t, uint64(18563), m.System)
	assert.Equal(t, uint64(1121456), m.Idle)
	assert.Equal(t, 4, m.NumCPUs)
	assert.Greater(t, m.UsagePct, 0.0)
	assert.Less(t, m.UsagePct, 100.0)
}

func TestParseMeminfo(t *testing.T) {
	content := `MemTotal:       16384000 kB
MemFree:         2048000 kB
MemAvailable:    8192000 kB
Buffers:          512000 kB
Cached:          4096000 kB
SwapTotal:       2048000 kB
SwapFree:        1024000 kB
`
	m := metrics.ParseMeminfo(content)
	assert.Equal(t, uint64(16384000), m.TotalKB)
	assert.Equal(t, uint64(2048000), m.FreeKB)
	assert.Equal(t, uint64(8192000), m.AvailableKB)
	assert.InDelta(t, 50.0, m.UsedPct, 0.1)
	assert.InDelta(t, 50.0, m.SwapUsedPct, 0.1)
}

func TestParseDF(t *testing.T) {
	content := `Filesystem     1K-blocks    Used Available Use% Mounted on
/dev/sda1      102400000 51200000  51200000  50% /
tmpfs            8192000        0   8192000   0% /dev/shm
/dev/sdb1       51200000 25600000  25600000  50% /data
`
	disks := metrics.ParseDF(content)
	require.Len(t, disks, 3)
	assert.Equal(t, "/dev/sda1", disks[0].Filesystem)
	assert.Equal(t, "/", disks[0].MountPoint)
	assert.Equal(t, 50.0, disks[0].UsedPct)
	assert.Equal(t, "/data", disks[2].MountPoint)
}

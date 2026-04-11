package metrics_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDiskStatsDelta(t *testing.T) {
	before := `   8       0 sda 12345 0 98760 5000 6789 0 54312 3000 0 4000 8000
   8      16 sdb 5000 0 40000 2000 3000 0 24000 1500 0 2000 3500
`
	after := `   8       0 sda 12445 0 99560 5100 6889 0 55112 3100 0 4100 8200
   8      16 sdb 5100 0 40800 2100 3100 0 24800 1600 0 2100 3700
`
	disks, err := metrics.ParseDiskStatsDelta(before, after, 1.0)
	require.NoError(t, err)
	assert.Len(t, disks, 2)

	sda := disks["sda"]
	assert.Greater(t, sda.ReadIOPS, 0.0)
	assert.Greater(t, sda.WriteIOPS, 0.0)
	assert.GreaterOrEqual(t, sda.ReadThroughputMB, 0.0)
	assert.GreaterOrEqual(t, sda.WriteThroughputMB, 0.0)
}

func TestParseDfMounts(t *testing.T) {
	content := `Filesystem     1K-blocks     Used Available Use% Mounted on
/dev/sda1       51474024 18234560  30600752  38% /
/dev/sdb1      103081248 42003456  55810080  43% /data
tmpfs            8192000        0   8192000   0% /dev/shm
`
	mounts, err := metrics.ParseDfMounts(content)
	require.NoError(t, err)
	assert.Len(t, mounts, 2)

	root := mounts["/"]
	assert.Equal(t, "/dev/sda1", root.Device)
	assert.Equal(t, uint64(51474024*1024), root.TotalBytes)
	assert.InDelta(t, 38.0, root.UsedPct, 1.0)
}

func TestParseProcNetDevDelta(t *testing.T) {
	before := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
  eth0: 1000000  10000    0    0    0     0          0         0  500000   5000    0    0    0     0       0          0
    lo:  200000   2000    0    0    0     0          0         0  200000   2000    0    0    0     0       0          0
`
	after := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
  eth0: 1100000  10500    0    0    0     0          0         0  600000   5500    0    0    0     0       0          0
    lo:  200500   2005    0    0    0     0          0         0  200500   2005    0    0    0     0       0          0
`
	nics, err := metrics.ParseNetDevDelta(before, after, 1.0)
	require.NoError(t, err)

	eth0 := nics["eth0"]
	assert.Equal(t, uint64(100000), eth0.RxBytesPerSec)
	assert.Equal(t, uint64(100000), eth0.TxBytesPerSec)
	assert.Equal(t, uint64(500), eth0.RxPacketsPerSec)
	assert.Equal(t, uint64(500), eth0.TxPacketsPerSec)
}

func TestParsePsAux(t *testing.T) {
	content := `USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
oracle     12345 42.0  8.2 4000000 2700000 ?   Sl   09:00  12:30 ora_dbw0_ORCL
oracle     12346 28.5  6.1 3500000 2000000 ?   Sl   09:00   8:15 ora_lgwr_ORCL
postgres   23456  5.2  3.1 1200000 1020000 ?   Ss   09:00   1:30 postgres: writer
root           1  0.0  0.1  168000  12000 ?   Ss   Mar01   0:05 /sbin/init
`
	procs, err := metrics.ParsePsAux(content)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(procs), 4)
	assert.Equal(t, "oracle", procs[0].User)
	assert.Equal(t, 12345, procs[0].PID)
	assert.InDelta(t, 42.0, procs[0].CPUPct, 0.1)
}

func TestTopN(t *testing.T) {
	procs := []metrics.ProcessInfo{
		{PID: 1, CPUPct: 42.0},
		{PID: 2, CPUPct: 28.5},
		{PID: 3, CPUPct: 5.2},
		{PID: 4, CPUPct: 0.0},
	}
	top2 := metrics.TopN(procs, 2, "cpu")
	assert.Len(t, top2, 2)
	assert.Equal(t, 1, top2[0].PID)
	assert.Equal(t, 2, top2[1].PID)
}

func TestCountZombies(t *testing.T) {
	procs := []metrics.ProcessInfo{
		{PID: 1, State: "S"},
		{PID: 2, State: "Z"},
		{PID: 3, State: "Z"},
		{PID: 4, State: "R"},
	}
	assert.Equal(t, 2, metrics.CountZombies(procs))
}

func TestParseLoadAvg(t *testing.T) {
	content := "0.42 0.51 0.63 2/1024 12345\n"
	load, err := metrics.ParseLoadAvg(content)
	require.NoError(t, err)
	assert.InDelta(t, 0.42, load.Load1, 0.001)
	assert.InDelta(t, 0.51, load.Load5, 0.001)
	assert.InDelta(t, 0.63, load.Load15, 0.001)
}

func TestParseLscpu(t *testing.T) {
	content := `Architecture:            x86_64
CPU(s):                  8
Thread(s) per core:      2
Core(s) per socket:      4
Socket(s):               1
Model name:              Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz
`
	info, err := metrics.ParseLscpu(content)
	require.NoError(t, err)
	assert.Equal(t, 8, info.CPUCount)
	assert.Equal(t, 4, info.CoresPerSocket)
	assert.Equal(t, 1, info.Sockets)
	assert.Contains(t, info.ModelName, "Xeon")
}

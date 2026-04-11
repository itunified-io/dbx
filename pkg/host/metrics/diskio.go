package metrics

import (
	"fmt"
	"strconv"
	"strings"
)

// DiskIOMetrics holds per-device I/O metrics computed from /proc/diskstats delta.
type DiskIOMetrics struct {
	Device           string  `json:"device"`
	ReadIOPS         float64 `json:"read_iops"`
	WriteIOPS        float64 `json:"write_iops"`
	ReadThroughputMB float64 `json:"read_throughput_mb"`
	WriteThroughputMB float64 `json:"write_throughput_mb"`
}

// ParseDiskStatsDelta computes per-device IOPS and throughput from two /proc/diskstats
// snapshots taken intervalSec apart.
func ParseDiskStatsDelta(before, after string, intervalSec float64) (map[string]DiskIOMetrics, error) {
	bMap, err := parseDiskStats(before)
	if err != nil {
		return nil, fmt.Errorf("parse before: %w", err)
	}
	aMap, err := parseDiskStats(after)
	if err != nil {
		return nil, fmt.Errorf("parse after: %w", err)
	}

	result := make(map[string]DiskIOMetrics)
	for dev, a := range aMap {
		b, ok := bMap[dev]
		if !ok {
			continue
		}
		result[dev] = DiskIOMetrics{
			Device:            dev,
			ReadIOPS:          float64(a.readsCompleted-b.readsCompleted) / intervalSec,
			WriteIOPS:         float64(a.writesCompleted-b.writesCompleted) / intervalSec,
			ReadThroughputMB:  float64(a.sectorsRead-b.sectorsRead) * 512 / 1024 / 1024 / intervalSec,
			WriteThroughputMB: float64(a.sectorsWritten-b.sectorsWritten) * 512 / 1024 / 1024 / intervalSec,
		}
	}
	return result, nil
}

type diskStatEntry struct {
	readsCompleted  uint64
	sectorsRead     uint64
	writesCompleted uint64
	sectorsWritten  uint64
}

func parseDiskStats(content string) (map[string]diskStatEntry, error) {
	result := make(map[string]diskStatEntry)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}
		dev := fields[2]
		reads, _ := strconv.ParseUint(fields[3], 10, 64)
		sectorsR, _ := strconv.ParseUint(fields[5], 10, 64)
		writes, _ := strconv.ParseUint(fields[7], 10, 64)
		sectorsW, _ := strconv.ParseUint(fields[9], 10, 64)
		result[dev] = diskStatEntry{
			readsCompleted:  reads,
			sectorsRead:     sectorsR,
			writesCompleted: writes,
			sectorsWritten:  sectorsW,
		}
	}
	return result, nil
}

// DfMount holds parsed df output with bytes instead of KB.
type DfMount struct {
	Device     string  `json:"device"`
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	AvailBytes uint64  `json:"avail_bytes"`
	UsedPct    float64 `json:"used_pct"`
	MountPoint string  `json:"mount_point"`
}

// ParseDfMounts parses `df -k` output, excludes tmpfs/devtmpfs, returns map by mount point.
func ParseDfMounts(content string) (map[string]DfMount, error) {
	result := make(map[string]DfMount)
	for i, line := range strings.Split(content, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		// Skip pseudo-filesystems
		if fields[0] == "tmpfs" || fields[0] == "devtmpfs" || fields[0] == "proc" || fields[0] == "sysfs" {
			continue
		}
		size, _ := strconv.ParseUint(fields[1], 10, 64)
		used, _ := strconv.ParseUint(fields[2], 10, 64)
		avail, _ := strconv.ParseUint(fields[3], 10, 64)
		pctStr := strings.TrimSuffix(fields[4], "%")
		pct, _ := strconv.ParseFloat(pctStr, 64)

		result[fields[5]] = DfMount{
			Device:     fields[0],
			TotalBytes: size * 1024,
			UsedBytes:  used * 1024,
			AvailBytes: avail * 1024,
			UsedPct:    pct,
			MountPoint: fields[5],
		}
	}
	return result, nil
}

package metrics

import (
	"strconv"
	"strings"
)

// DiskMetrics holds parsed df output for a single filesystem.
type DiskMetrics struct {
	Filesystem string  `json:"filesystem"`
	SizeKB     uint64  `json:"size_kb"`
	UsedKB     uint64  `json:"used_kb"`
	AvailKB    uint64  `json:"avail_kb"`
	UsedPct    float64 `json:"used_pct"`
	MountPoint string  `json:"mount_point"`
}

// ParseDF parses the output of `df -k` (POSIX output) and returns disk metrics.
func ParseDF(content string) []DiskMetrics {
	var result []DiskMetrics
	for i, line := range strings.Split(content, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		size, _ := strconv.ParseUint(fields[1], 10, 64)
		used, _ := strconv.ParseUint(fields[2], 10, 64)
		avail, _ := strconv.ParseUint(fields[3], 10, 64)
		pctStr := strings.TrimSuffix(fields[4], "%")
		pct, _ := strconv.ParseFloat(pctStr, 64)

		result = append(result, DiskMetrics{
			Filesystem: fields[0],
			SizeKB:     size,
			UsedKB:     used,
			AvailKB:    avail,
			UsedPct:    pct,
			MountPoint: fields[5],
		})
	}
	return result
}

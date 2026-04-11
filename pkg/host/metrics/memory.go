package metrics

import (
	"strconv"
	"strings"
)

// MemoryMetrics holds parsed /proc/meminfo values.
type MemoryMetrics struct {
	TotalKB     uint64  `json:"total_kb"`
	FreeKB      uint64  `json:"free_kb"`
	AvailableKB uint64  `json:"available_kb"`
	BuffersKB   uint64  `json:"buffers_kb"`
	CachedKB    uint64  `json:"cached_kb"`
	SwapTotalKB uint64  `json:"swap_total_kb"`
	SwapFreeKB  uint64  `json:"swap_free_kb"`
	UsedPct     float64 `json:"used_pct"`
	SwapUsedPct float64 `json:"swap_used_pct"`
}

// ParseMeminfo parses /proc/meminfo content and returns memory metrics.
func ParseMeminfo(content string) *MemoryMetrics {
	m := &MemoryMetrics{}

	for _, line := range strings.Split(content, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, _ := strconv.ParseUint(parts[1], 10, 64)

		switch key {
		case "MemTotal":
			m.TotalKB = val
		case "MemFree":
			m.FreeKB = val
		case "MemAvailable":
			m.AvailableKB = val
		case "Buffers":
			m.BuffersKB = val
		case "Cached":
			m.CachedKB = val
		case "SwapTotal":
			m.SwapTotalKB = val
		case "SwapFree":
			m.SwapFreeKB = val
		}
	}

	if m.TotalKB > 0 {
		used := m.TotalKB - m.AvailableKB
		m.UsedPct = float64(used) / float64(m.TotalKB) * 100
	}
	if m.SwapTotalKB > 0 {
		swapUsed := m.SwapTotalKB - m.SwapFreeKB
		m.SwapUsedPct = float64(swapUsed) / float64(m.SwapTotalKB) * 100
	}

	return m
}

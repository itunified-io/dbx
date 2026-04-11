package metrics

import (
	"sort"
	"strconv"
	"strings"
)

// ProcessInfo holds a parsed process from `ps aux` output.
type ProcessInfo struct {
	User    string  `json:"user"`
	PID     int     `json:"pid"`
	CPUPct  float64 `json:"cpu_pct"`
	MemPct  float64 `json:"mem_pct"`
	VSZ     uint64  `json:"vsz"`
	RSS     uint64  `json:"rss"`
	State   string  `json:"state"`
	Command string  `json:"command"`
}

// ParsePsAux parses `ps aux --sort=-%cpu` output.
func ParsePsAux(content string) ([]ProcessInfo, error) {
	var procs []ProcessInfo
	for i, line := range strings.Split(content, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, _ := strconv.Atoi(fields[1])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		vsz, _ := strconv.ParseUint(fields[4], 10, 64)
		rss, _ := strconv.ParseUint(fields[5], 10, 64)

		procs = append(procs, ProcessInfo{
			User:    fields[0],
			PID:     pid,
			CPUPct:  cpu,
			MemPct:  mem,
			VSZ:     vsz,
			RSS:     rss,
			State:   fields[7],
			Command: strings.Join(fields[10:], " "),
		})
	}
	return procs, nil
}

// TopN returns the top N processes sorted by the given field ("cpu" or "mem").
func TopN(procs []ProcessInfo, n int, sortBy string) []ProcessInfo {
	sorted := make([]ProcessInfo, len(procs))
	copy(sorted, procs)

	sort.Slice(sorted, func(i, j int) bool {
		if sortBy == "mem" {
			return sorted[i].MemPct > sorted[j].MemPct
		}
		return sorted[i].CPUPct > sorted[j].CPUPct
	})

	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

// CountZombies returns the number of zombie processes.
func CountZombies(procs []ProcessInfo) int {
	count := 0
	for _, p := range procs {
		if p.State == "Z" {
			count++
		}
	}
	return count
}

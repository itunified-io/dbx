// Package metrics provides system metric collectors for CPU, memory, disk, network, and processes.
package metrics

import (
	"fmt"
	"strconv"
	"strings"
)

// CPUMetrics holds parsed /proc/stat values.
type CPUMetrics struct {
	User      uint64  `json:"user"`
	Nice      uint64  `json:"nice"`
	System    uint64  `json:"system"`
	Idle      uint64  `json:"idle"`
	IOWait    uint64  `json:"iowait"`
	IRQ       uint64  `json:"irq"`
	SoftIRQ   uint64  `json:"softirq"`
	Steal     uint64  `json:"steal"`
	NumCPUs   int     `json:"num_cpus"`
	UsagePct  float64 `json:"usage_pct"`
}

// ParseProcStat parses /proc/stat content and returns CPU metrics.
func ParseProcStat(content string) (*CPUMetrics, error) {
	m := &CPUMetrics{}

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return nil, fmt.Errorf("unexpected /proc/stat format: too few fields")
			}
			var err error
			m.User, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse user: %w", err)
			}
			m.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
			m.System, _ = strconv.ParseUint(fields[3], 10, 64)
			m.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
			m.IOWait, _ = strconv.ParseUint(fields[5], 10, 64)
			m.IRQ, _ = strconv.ParseUint(fields[6], 10, 64)
			m.SoftIRQ, _ = strconv.ParseUint(fields[7], 10, 64)
			if len(fields) > 8 {
				m.Steal, _ = strconv.ParseUint(fields[8], 10, 64)
			}
		}
		if strings.HasPrefix(line, "cpu") && !strings.HasPrefix(line, "cpu ") {
			m.NumCPUs++
		}
	}

	total := m.User + m.Nice + m.System + m.Idle + m.IOWait + m.IRQ + m.SoftIRQ + m.Steal
	if total > 0 {
		busy := total - m.Idle - m.IOWait
		m.UsagePct = float64(busy) / float64(total) * 100
	}

	return m, nil
}

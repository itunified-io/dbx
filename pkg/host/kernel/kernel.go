// Package kernel provides kernel parameter, module, and hugepage parsing.
package kernel

import (
	"strconv"
	"strings"
)

// ParseSysctl parses the output of `sysctl -a`.
func ParseSysctl(output string) (map[string]string, error) {
	params := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return params, nil
}

// Hugepages holds parsed hugepage configuration from /proc/meminfo.
type Hugepages struct {
	Total     uint64  `json:"total"`
	Free      uint64  `json:"free"`
	Reserved  uint64  `json:"reserved"`
	Surplus   uint64  `json:"surplus"`
	SizeBytes uint64  `json:"size_bytes"`
	UsedPct   float64 `json:"used_pct"`
}

// ParseHugepages parses hugepage-related lines from /proc/meminfo.
func ParseHugepages(content string) (*Hugepages, error) {
	hp := &Hugepages{}
	for _, line := range strings.Split(content, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, _ := strconv.ParseUint(parts[1], 10, 64)

		switch key {
		case "HugePages_Total":
			hp.Total = val
		case "HugePages_Free":
			hp.Free = val
		case "HugePages_Rsvd":
			hp.Reserved = val
		case "HugePages_Surp":
			hp.Surplus = val
		case "Hugepagesize":
			hp.SizeBytes = val * 1024
		}
	}
	if hp.Total > 0 {
		used := hp.Total - hp.Free
		hp.UsedPct = float64(used) / float64(hp.Total) * 100
	}
	return hp, nil
}

// KernelInfo holds parsed uname -r output.
type KernelInfo struct {
	Release string `json:"release"`
	Version string `json:"version"` // Major.minor.patch extracted
}

// ParseUnameR parses the output of `uname -r`.
func ParseUnameR(output string) KernelInfo {
	release := strings.TrimSpace(output)
	version := release
	// Extract version up to first non-numeric.non-numeric.non-numeric
	parts := strings.SplitN(release, "-", 2)
	if len(parts) >= 1 {
		version = parts[0]
	}
	return KernelInfo{Release: release, Version: version}
}

// Module represents a loaded kernel module from lsmod output.
type Module struct {
	Name   string `json:"name"`
	Size   uint64 `json:"size"`
	UsedBy int    `json:"used_by"`
}

// ParseLsmod parses the output of `lsmod`.
func ParseLsmod(output string) ([]Module, error) {
	var modules []Module
	for i, line := range strings.Split(output, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		size, _ := strconv.ParseUint(fields[1], 10, 64)
		usedBy, _ := strconv.Atoi(fields[2])
		modules = append(modules, Module{
			Name:   fields[0],
			Size:   size,
			UsedBy: usedBy,
		})
	}
	return modules, nil
}

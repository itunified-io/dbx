// Package filesystem provides mount point, LVM, and inode parsing.
package filesystem

import (
	"strconv"
	"strings"
)

// Mount represents a parsed mount point from findmnt output.
type Mount struct {
	Target  string `json:"target"`
	Source  string `json:"source"`
	FSType  string `json:"fstype"`
	Options string `json:"options"`
}

// ParseFindmnt parses `findmnt -l -n -o TARGET,SOURCE,FSTYPE,OPTIONS` output.
func ParseFindmnt(output string) ([]Mount, error) {
	var mounts []Mount
	for i, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip header
		if i == 0 && strings.HasPrefix(line, "TARGET") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		m := Mount{
			Target: fields[0],
			Source: fields[1],
			FSType: fields[2],
		}
		if len(fields) >= 4 {
			m.Options = fields[3]
		}
		mounts = append(mounts, m)
	}
	return mounts, nil
}

// PV represents a physical volume from pvs output.
type PV struct {
	Device string `json:"device"`
	VG     string `json:"vg"`
	Format string `json:"format"`
	Size   string `json:"size"`
	Free   string `json:"free"`
}

// LV represents a logical volume from lvs output.
type LV struct {
	Name string `json:"name"`
	VG   string `json:"vg"`
	Attr string `json:"attr"`
	Size string `json:"size"`
}

// LVMLayout holds parsed LVM topology.
type LVMLayout struct {
	PVs []PV `json:"pvs"`
	LVs []LV `json:"lvs"`
}

// ParseLVMLayout parses pvs and lvs output.
func ParseLVMLayout(pvsOutput, lvsOutput string) (*LVMLayout, error) {
	layout := &LVMLayout{}

	for _, line := range strings.Split(pvsOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "PV") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		layout.PVs = append(layout.PVs, PV{
			Device: fields[0],
			VG:     fields[1],
			Format: fields[2],
			Size:   fields[4],
			Free:   func() string {
				if len(fields) >= 6 {
					return fields[5]
				}
				return ""
			}(),
		})
	}

	for _, line := range strings.Split(lvsOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "LV") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		layout.LVs = append(layout.LVs, LV{
			Name: fields[0],
			VG:   fields[1],
			Attr: fields[2],
			Size: fields[3],
		})
	}

	return layout, nil
}

// InodeInfo holds inode counts for a mount point.
type InodeInfo struct {
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	UsedPct float64 `json:"used_pct"`
}

// ParseDfInodes parses `df -i` output and returns inode info by mount point.
func ParseDfInodes(output string) (map[string]InodeInfo, error) {
	result := make(map[string]InodeInfo)
	for i, line := range strings.Split(output, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		total, _ := strconv.ParseUint(fields[1], 10, 64)
		used, _ := strconv.ParseUint(fields[2], 10, 64)
		free, _ := strconv.ParseUint(fields[3], 10, 64)
		pctStr := strings.TrimSuffix(fields[4], "%")
		pct, _ := strconv.ParseFloat(pctStr, 64)

		mountPoint := fields[5]
		result[mountPoint] = InodeInfo{
			Total:   total,
			Used:    used,
			Free:    free,
			UsedPct: pct,
		}
	}
	return result, nil
}

// Package patch provides OS patch status and ksplice assessment.
package patch

import (
	"strings"
	"time"
)

// PatchStatus holds aggregated patch/update statistics.
type PatchStatus struct {
	TotalUpdates    int `json:"total_updates"`
	SecurityUpdates int `json:"security_updates"`
	CriticalCount   int `json:"critical_count"`
	ImportantCount  int `json:"important_count"`
	ModerateCount   int `json:"moderate_count"`
}

// ParsePatchStatus aggregates update info by severity from dnf updateinfo output.
func ParsePatchStatus(output string) (*PatchStatus, error) {
	ps := &PatchStatus{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ps.TotalUpdates++
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		severity := fields[1]
		if strings.Contains(severity, "Sec.") || strings.Contains(severity, "/Sec") {
			ps.SecurityUpdates++
			sev := strings.SplitN(severity, "/", 2)[0]
			switch sev {
			case "Critical":
				ps.CriticalCount++
			case "Important":
				ps.ImportantCount++
			case "Moderate":
				ps.ModerateCount++
			}
		}
	}
	return ps, nil
}

// KsplicePatch represents an installed ksplice live patch.
type KsplicePatch struct {
	CVE      string `json:"cve"`
	Severity string `json:"severity"`
}

// KspliceStatus holds parsed ksplice status.
type KspliceStatus struct {
	EffectiveKernel  string         `json:"effective_kernel"`
	InstalledPatches []KsplicePatch `json:"installed_patches"`
}

// ParseKspliceStatus parses `ksplice show` output.
func ParseKspliceStatus(output string) (*KspliceStatus, error) {
	ks := &KspliceStatus{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Effective kernel version:") {
			ks.EffectiveKernel = strings.TrimSpace(strings.TrimPrefix(line, "Effective kernel version:"))
		}
		if strings.HasPrefix(line, "CVE-") {
			parts := strings.Fields(line)
			patch := KsplicePatch{CVE: parts[0]}
			if len(parts) >= 2 {
				patch.Severity = strings.Trim(parts[1], "[]")
			}
			ks.InstalledPatches = append(ks.InstalledPatches, patch)
		}
	}
	return ks, nil
}

// ParseLastUpdateTime parses `rpm -q --last | head -1` to extract last update timestamp.
func ParseLastUpdateTime(output string) (time.Time, error) {
	line := strings.TrimSpace(output)
	if line == "" {
		return time.Time{}, nil
	}
	// Format: "pkg-name  Day DD Mon YYYY HH:MM:SS AM/PM TZ"
	// Find the first double-space separator between package and date
	parts := strings.SplitN(line, "  ", 2)
	if len(parts) < 2 {
		return time.Time{}, nil
	}
	dateStr := strings.TrimSpace(parts[1])
	// Try common RPM date formats
	formats := []string{
		"Mon 02 Jan 2006 03:04:05 PM MST",
		"Mon Jan 2 15:04:05 2006",
		"Mon 02 Jan 2006 15:04:05 MST",
	}
	for _, fmt := range formats {
		if t, err := time.Parse(fmt, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}

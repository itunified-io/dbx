// Package hostpkg provides distro-agnostic package management parsing (apt/dnf/zypper).
package hostpkg

import (
	"strings"

	"github.com/itunified-io/dbx/pkg/host/distro"
)

// Package represents an installed OS package.
type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
}

// Update represents an available package update.
type Update struct {
	Name       string `json:"name"`
	NewVersion string `json:"new_version"`
	OldVersion string `json:"old_version"`
	Severity   string `json:"severity,omitempty"`
	Security   bool   `json:"security"`
	Advisory   string `json:"advisory,omitempty"`
}

// ParsePackageList dispatches to the appropriate parser based on distro.
func ParsePackageList(output string, id distro.DistroID) ([]Package, error) {
	switch id {
	case distro.Ubuntu:
		return parseDpkgList(output)
	default:
		return parseRpmList(output)
	}
}

// ParseSecurityUpdates dispatches to the appropriate parser based on distro.
func ParseSecurityUpdates(output string, id distro.DistroID) ([]Update, error) {
	switch id {
	case distro.Ubuntu:
		return parseAptUpdates(output)
	default:
		return parseDnfUpdateInfo(output)
	}
}

func parseRpmList(output string) ([]Package, error) {
	var pkgs []Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 2 {
			continue
		}
		pkg := Package{
			Name:    strings.TrimSpace(parts[0]),
			Version: strings.TrimSpace(parts[1]),
		}
		if len(parts) >= 3 {
			pkg.Arch = strings.TrimSpace(parts[2])
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func parseDpkgList(output string) ([]Package, error) {
	var pkgs []Package
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		// Only include installed packages (status "ii")
		if fields[0] != "ii" {
			continue
		}
		name := fields[1]
		// Strip architecture suffix (e.g., libc6:amd64 -> libc6)
		if idx := strings.Index(name, ":"); idx > 0 {
			name = name[:idx]
		}
		pkgs = append(pkgs, Package{
			Name:    name,
			Version: fields[2],
			Arch:    fields[3],
		})
	}
	return pkgs, nil
}

func parseDnfUpdateInfo(output string) ([]Update, error) {
	var updates []Update
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		advisory := fields[0]
		severityField := fields[1]
		pkg := fields[2]

		severity := ""
		security := false
		if strings.Contains(severityField, "/Sec.") {
			security = true
			severity = strings.SplitN(severityField, "/", 2)[0]
		} else {
			severity = severityField
		}

		updates = append(updates, Update{
			Name:       pkg,
			Advisory:   advisory,
			Severity:   severity,
			Security:   security,
			NewVersion: pkg,
		})
	}
	return updates, nil
}

func parseAptUpdates(output string) ([]Update, error) {
	var updates []Update
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "name/source version arch [upgradable from: old_version]"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		nameSource := strings.SplitN(fields[0], "/", 2)
		name := nameSource[0]
		newVer := fields[1]

		oldVer := ""
		if idx := strings.Index(line, "upgradable from:"); idx >= 0 {
			rest := line[idx+len("upgradable from:"):]
			rest = strings.TrimSuffix(strings.TrimSpace(rest), "]")
			oldVer = strings.TrimSpace(rest)
		}

		updates = append(updates, Update{
			Name:       name,
			NewVersion: newVer,
			OldVersion: oldVer,
		})
	}
	return updates, nil
}

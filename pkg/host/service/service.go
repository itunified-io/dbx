// Package service provides systemd service management parsing.
package service

import (
	"strings"
)

// Unit represents a parsed systemd service unit.
type Unit struct {
	Name        string `json:"name"`
	LoadState   string `json:"load_state"`   // loaded, not-found, masked
	ActiveState string `json:"active_state"` // active, inactive, failed
	SubState    string `json:"sub_state"`    // running, dead, exited, failed
	Description string `json:"description"`
}

// ParseSystemctlList parses the output of `systemctl list-units --type=service --all --no-pager`.
func ParseSystemctlList(content string) []Unit {
	var units []Unit
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "UNIT") || strings.HasPrefix(line, "LOAD") {
			continue
		}
		// Skip summary/legend lines
		if strings.Contains(line, "loaded units listed") || strings.Contains(line, "To show all") ||
			strings.HasPrefix(line, "LOAD") || strings.HasPrefix(line, "ACTIVE") ||
			strings.HasPrefix(line, "SUB") || !strings.Contains(line, ".service") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Sometimes there's a dot prefix for inactive units
		name := fields[0]
		if strings.HasPrefix(name, "\u25cf") || strings.HasPrefix(name, "●") {
			if len(fields) < 5 {
				continue
			}
			name = fields[1]
			fields = fields[1:]
		}

		desc := ""
		if len(fields) > 4 {
			desc = strings.Join(fields[4:], " ")
		}

		units = append(units, Unit{
			Name:        name,
			LoadState:   fields[1],
			ActiveState: fields[2],
			SubState:    fields[3],
			Description: desc,
		})
	}
	return units
}

// FailedUnits returns only units in failed state.
func FailedUnits(units []Unit) []Unit {
	var failed []Unit
	for _, u := range units {
		if u.ActiveState == "failed" {
			failed = append(failed, u)
		}
	}
	return failed
}

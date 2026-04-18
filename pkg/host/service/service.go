// Package service provides systemd service management parsing.
package service

import (
	"fmt"
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

// FilterByState returns units matching the given active state.
func FilterByState(units []Unit, state string) []Unit {
	var result []Unit
	for _, u := range units {
		if u.ActiveState == state {
			result = append(result, u)
		}
	}
	return result
}

// ServiceDetail holds parsed `systemctl show` output.
type ServiceDetail struct {
	Type          string `json:"type"`
	ActiveState   string `json:"active_state"`
	SubState      string `json:"sub_state"`
	MainPID       int    `json:"main_pid"`
	MemoryBytes   uint64 `json:"memory_bytes"`
	StartTime     string `json:"start_time"`
	RestartPolicy string `json:"restart_policy"`
}

// ParseSystemctlShow parses `systemctl show <service>` output.
func ParseSystemctlShow(output string) (*ServiceDetail, error) {
	d := &ServiceDetail{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		switch key {
		case "Type":
			d.Type = val
		case "ActiveState":
			d.ActiveState = val
		case "SubState":
			d.SubState = val
		case "MainPID":
			fmt.Sscan(val, &d.MainPID)
		case "MemoryCurrent":
			fmt.Sscan(val, &d.MemoryBytes)
		case "ExecMainStartTimestamp":
			d.StartTime = val
		case "Restart":
			d.RestartPolicy = val
		}
	}
	return d, nil
}

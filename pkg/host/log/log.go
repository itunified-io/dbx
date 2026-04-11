// Package log provides journald and auth log analysis.
package log

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Entry represents a parsed journal entry.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Unit      string `json:"unit"`
	Priority  int    `json:"priority"` // 0=emerg, 3=err, 4=warning, 6=info, 7=debug
	Message   string `json:"message"`
	Identifier string `json:"identifier"`
}

// ParseJournalJSON parses `journalctl -o json` output (one JSON object per line).
func ParseJournalJSON(output string) ([]Entry, error) {
	var entries []Entry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		priority := 6
		if p, ok := raw["PRIORITY"].(string); ok {
			priority, _ = strconv.Atoi(p)
		}
		entries = append(entries, Entry{
			Timestamp:  getString(raw, "__REALTIME_TIMESTAMP"),
			Unit:       getString(raw, "_SYSTEMD_UNIT"),
			Priority:   priority,
			Message:    getString(raw, "MESSAGE"),
			Identifier: getString(raw, "SYSLOG_IDENTIFIER"),
		})
	}
	return entries, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// AuthLogSummary holds aggregated auth log statistics.
type AuthLogSummary struct {
	SuccessCount  int            `json:"success_count"`
	FailedCount   int            `json:"failed_count"`
	FailedSources map[string]int `json:"failed_sources"`
}

// ParseAuthLog parses /var/log/auth.log or /var/log/secure for SSH events.
func ParseAuthLog(output string) (*AuthLogSummary, error) {
	summary := &AuthLogSummary{
		FailedSources: make(map[string]int),
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "Accepted") {
			summary.SuccessCount++
		} else if strings.Contains(line, "Failed") {
			summary.FailedCount++
			// Extract source IP
			if idx := strings.Index(line, "from "); idx >= 0 {
				rest := line[idx+5:]
				fields := strings.Fields(rest)
				if len(fields) > 0 {
					summary.FailedSources[fields[0]]++
				}
			}
		}
	}
	return summary, nil
}

// FilterBySeverity returns entries with priority <= maxPriority (lower = more severe).
func FilterBySeverity(entries []Entry, maxPriority int) []Entry {
	var result []Entry
	for _, e := range entries {
		if e.Priority <= maxPriority {
			result = append(result, e)
		}
	}
	return result
}

// FilterByUnit returns entries matching the given systemd unit.
func FilterByUnit(entries []Entry, unit string) []Entry {
	var result []Entry
	for _, e := range entries {
		if e.Unit == unit {
			result = append(result, e)
		}
	}
	return result
}

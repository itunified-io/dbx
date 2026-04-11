// Package user provides user account auditing — who logged in, when, from where.
package user

import (
	"strings"
)

// LoginRecord represents a parsed `last` output line.
type LoginRecord struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Source   string `json:"source"`
	Login    string `json:"login"`
	Logout   string `json:"logout"`
	Duration string `json:"duration"`
}

// ParseLast parses the output of the `last` command.
func ParseLast(content string) []LoginRecord {
	var records []LoginRecord
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "wtmp") || strings.HasPrefix(line, "reboot") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		r := LoginRecord{
			User:     fields[0],
			Terminal: fields[1],
		}

		// Determine if source IP/hostname is present
		if len(fields) >= 10 {
			r.Source = fields[2]
			r.Login = strings.Join(fields[3:7], " ")
		} else if len(fields) >= 7 {
			r.Login = strings.Join(fields[2:6], " ")
		}

		records = append(records, r)
	}
	return records
}

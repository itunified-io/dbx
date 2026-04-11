// Package user provides user account auditing — who logged in, when, from where.
package user

import (
	"strconv"
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

// OSUser represents a parsed /etc/passwd entry.
type OSUser struct {
	Name          string `json:"name"`
	UID           int    `json:"uid"`
	GID           int    `json:"gid"`
	Comment       string `json:"comment"`
	Home          string `json:"home"`
	Shell         string `json:"shell"`
	HasLoginShell bool   `json:"has_login_shell"`
}

// ParsePasswd parses /etc/passwd content.
func ParsePasswd(content string) ([]OSUser, error) {
	var users []OSUser
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}
		uid, _ := strconv.Atoi(parts[2])
		gid, _ := strconv.Atoi(parts[3])
		shell := parts[6]
		hasLogin := !strings.Contains(shell, "nologin") && !strings.Contains(shell, "/false")
		users = append(users, OSUser{
			Name:          parts[0],
			UID:           uid,
			GID:           gid,
			Comment:       parts[4],
			Home:          parts[5],
			Shell:         shell,
			HasLoginShell: hasLogin,
		})
	}
	return users, nil
}

// Group represents a parsed /etc/group entry.
type Group struct {
	Name    string   `json:"name"`
	GID     int      `json:"gid"`
	Members []string `json:"members"`
}

// ParseGroups parses /etc/group content.
func ParseGroups(content string) ([]Group, error) {
	var groups []Group
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}
		gid, _ := strconv.Atoi(parts[2])
		var members []string
		if parts[3] != "" {
			members = strings.Split(parts[3], ",")
		}
		groups = append(groups, Group{
			Name:    parts[0],
			GID:     gid,
			Members: members,
		})
	}
	return groups, nil
}

// Session represents a parsed `who` output entry.
type Session struct {
	User       string `json:"user"`
	Terminal   string `json:"terminal"`
	LoginTime  string `json:"login_time"`
	RemoteHost string `json:"remote_host"`
}

// ParseWho parses the output of `who`.
func ParseWho(output string) ([]Session, error) {
	var sessions []Session
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		s := Session{
			User:     fields[0],
			Terminal: fields[1],
		}
		if len(fields) >= 4 {
			s.LoginTime = fields[2] + " " + fields[3]
		}
		// Remote host is in parentheses
		if idx := strings.Index(line, "("); idx >= 0 {
			end := strings.Index(line[idx:], ")")
			if end > 0 {
				s.RemoteHost = line[idx+1 : idx+end]
			}
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// SudoersRule represents a parsed sudoers entry.
type SudoersRule struct {
	Principal string   `json:"principal"` // user or %group
	Host      string   `json:"host"`
	RunAs     string   `json:"run_as"`
	NOPASSWD  bool     `json:"nopasswd"`
	Commands  []string `json:"commands"`
}

// ParseSudoers parses /etc/sudoers content.
func ParseSudoers(content string) ([]SudoersRule, error) {
	var rules []SudoersRule
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "Defaults") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		rule := SudoersRule{Principal: fields[0]}
		rest := strings.Join(fields[1:], " ")

		// Parse host=(runas)
		if idx := strings.Index(rest, "("); idx >= 0 {
			rule.Host = strings.TrimSpace(rest[:idx])
			end := strings.Index(rest[idx:], ")")
			if end > 0 {
				rule.RunAs = rest[idx+1 : idx+end]
			}
			rest = strings.TrimSpace(rest[idx+end+1:])
		}

		// Check for NOPASSWD
		if strings.HasPrefix(rest, "NOPASSWD:") {
			rule.NOPASSWD = true
			rest = strings.TrimSpace(strings.TrimPrefix(rest, "NOPASSWD:"))
		} else if strings.HasPrefix(rest, "ALL") {
			rest = strings.TrimSpace(rest)
		}

		// Parse commands
		for _, cmd := range strings.Split(rest, ",") {
			cmd = strings.TrimSpace(cmd)
			if cmd != "" {
				rule.Commands = append(rule.Commands, cmd)
			}
		}

		rules = append(rules, rule)
	}
	return rules, nil
}

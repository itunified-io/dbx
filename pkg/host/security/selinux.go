// Package security provides security posture assessment — SELinux, AppArmor, compliance checks.
package security

import "strings"

// SELinuxStatus holds parsed sestatus output.
type SELinuxStatus struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"` // enforcing, permissive, disabled
	Policy  string `json:"policy"`
}

// ParseSEStatus parses the output of `sestatus`.
func ParseSEStatus(content string) *SELinuxStatus {
	s := &SELinuxStatus{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch {
		case strings.Contains(key, "SELinux status"):
			s.Enabled = val == "enabled"
		case strings.Contains(key, "Current mode"):
			s.Mode = val
		case strings.Contains(key, "Loaded policy"):
			s.Policy = val
		case strings.Contains(key, "Mode from config"):
			if s.Mode == "" {
				s.Mode = val
			}
		}
	}
	return s
}

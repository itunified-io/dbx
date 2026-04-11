// Package security provides security posture assessment — SELinux, AppArmor, SSH, compliance checks.
package security

import (
	"strconv"
	"strings"
)

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

// ParseSELinuxStatus is an alias for ParseSEStatus with error return for consistency.
func ParseSELinuxStatus(content string) (*SELinuxStatus, error) {
	return ParseSEStatus(content), nil
}

// AppArmorStatus holds parsed aa-status output.
type AppArmorStatus struct {
	Loaded          bool `json:"loaded"`
	ProfilesEnforce int  `json:"profiles_enforce"`
	ProfilesComplain int `json:"profiles_complain"`
	Unconfined      int  `json:"unconfined"`
}

// ParseAppArmorStatus parses `aa-status` output.
func ParseAppArmorStatus(content string) (*AppArmorStatus, error) {
	aa := &AppArmorStatus{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "apparmor module is loaded") {
			aa.Loaded = true
		}
		if strings.Contains(line, "profiles are in enforce mode") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				aa.ProfilesEnforce, _ = strconv.Atoi(fields[0])
			}
		}
		if strings.Contains(line, "profiles are in complain mode") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				aa.ProfilesComplain, _ = strconv.Atoi(fields[0])
			}
		}
		if strings.Contains(line, "processes are unconfined") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				aa.Unconfined, _ = strconv.Atoi(fields[0])
			}
		}
	}
	return aa, nil
}

// SSHDConfig holds SSH daemon security checks.
type SSHDConfig struct {
	RootLoginDisabled    bool `json:"root_login_disabled"`
	PasswordAuthDisabled bool `json:"password_auth_disabled"`
	PubkeyEnabled        bool `json:"pubkey_enabled"`
	MaxAuthTries         int  `json:"max_auth_tries"`
	X11Forwarding        bool `json:"x11_forwarding"`
}

// CheckSSHDConfig validates /etc/ssh/sshd_config against security best practices.
func CheckSSHDConfig(content string) SSHDConfig {
	cfg := SSHDConfig{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(parts[0])
		val := strings.ToLower(parts[1])
		switch key {
		case "permitrootlogin":
			cfg.RootLoginDisabled = val == "no"
		case "passwordauthentication":
			cfg.PasswordAuthDisabled = val == "no"
		case "pubkeyauthentication":
			cfg.PubkeyEnabled = val == "yes"
		case "maxauthtries":
			cfg.MaxAuthTries, _ = strconv.Atoi(parts[1])
		case "x11forwarding":
			cfg.X11Forwarding = val == "yes"
		}
	}
	return cfg
}

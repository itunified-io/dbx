// Package network provides Oracle Linux network diagnostics over SSH.
package network

import (
	"github.com/itunified-io/dbx/pkg/core/ssh"
)

// Manager provides network diagnostic operations (all read-only).
type Manager struct {
	ssh *ssh.Executor
}

// New creates a network manager with the given SSH executor.
func New(exec *ssh.Executor) *Manager {
	return &Manager{ssh: exec}
}

// NicListArgs builds SSH args for listing NICs with addresses (JSON output).
func (m *Manager) NicListArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "ip", []string{"-j", "addr", "show"})
}

// BondStatusArgs builds SSH args for showing network bond/connection status.
func (m *Manager) BondStatusArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "nmcli", []string{"connection", "show"})
}

// DnsCheckArgs builds SSH args for reading DNS resolver configuration.
func (m *Manager) DnsCheckArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "cat", []string{"/etc/resolv.conf"})
}

// NtpStatusArgs builds SSH args for checking NTP synchronization status.
func (m *Manager) NtpStatusArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "chronyc", []string{"tracking"})
}

// Package security provides Oracle Linux security status checks over SSH.
package security

import (
	"github.com/itunified-io/dbx/pkg/core/ssh"
)

// Manager provides security status operations (all read-only).
type Manager struct {
	ssh *ssh.Executor
}

// New creates a security manager with the given SSH executor.
func New(exec *ssh.Executor) *Manager {
	return &Manager{ssh: exec}
}

// SelinuxStatusArgs builds SSH args for checking SELinux status.
func (m *Manager) SelinuxStatusArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "sestatus", nil)
}

// FirewallListArgs builds SSH args for listing firewall rules.
func (m *Manager) FirewallListArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "firewall-cmd", []string{"--list-all"})
}

// ServiceStatusArgs builds SSH args for listing running services.
func (m *Manager) ServiceStatusArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "systemctl",
		[]string{"list-units", "--type=service", "--state=running", "--no-pager"})
}

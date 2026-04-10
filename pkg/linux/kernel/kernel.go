// Package kernel provides Oracle Linux kernel parameter management over SSH.
package kernel

import (
	"fmt"
	"strings"

	"github.com/itunified-io/dbx/pkg/core/ssh"
)

// Manager provides kernel parameter operations.
type Manager struct {
	ssh *ssh.Executor
}

// New creates a kernel manager with the given SSH executor.
func New(exec *ssh.Executor) *Manager {
	return &Manager{ssh: exec}
}

// ParamListArgs builds SSH args for listing all sysctl parameters.
func (m *Manager) ParamListArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "sysctl", []string{"-a"})
}

// ParamSetArgs builds SSH args for setting a sysctl parameter (confirm-gated).
func (m *Manager) ParamSetArgs(user, host, keyPath, key, value string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "sysctl", []string{"-w", fmt.Sprintf("%s=%s", key, value)})
}

// HugepagesSetArgs builds SSH args for setting hugepages count (confirm-gated).
func (m *Manager) HugepagesSetArgs(user, host, keyPath string, pages int) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "sysctl", []string{"-w", fmt.Sprintf("vm.nr_hugepages=%d", pages)})
}

// InfoArgs builds SSH args for OS kernel info.
func (m *Manager) InfoArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "uname", []string{"-a"})
}

// ParseSysctl parses `key = value` lines into a map.
func ParseSysctl(raw string) map[string]string {
	params := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return params
}

// IsMutating reports whether the operation modifies kernel state.
func IsMutating(op string) bool {
	return op == "param_set" || op == "hugepages"
}

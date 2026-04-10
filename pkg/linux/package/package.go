// Package linuxpkg provides Oracle Linux package management (RPM/DNF) over SSH.
package linuxpkg

import (
	"strings"

	"github.com/itunified-io/dbx/pkg/core/ssh"
)

// Package represents an installed RPM package.
type Package struct {
	Name    string
	Version string
	Arch    string
}

// Manager provides package management operations.
type Manager struct {
	ssh *ssh.Executor
}

// New creates a package manager with the given SSH executor.
func New(exec *ssh.Executor) *Manager {
	return &Manager{ssh: exec}
}

// ListArgs builds SSH args for listing all installed packages.
func (m *Manager) ListArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "rpm",
		[]string{"-qa", "--queryformat", "%{NAME}|%{VERSION}-%{RELEASE}|%{ARCH}\\n"})
}

// InfoArgs builds SSH args for getting package info.
func (m *Manager) InfoArgs(user, host, keyPath, pkg string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "rpm", []string{"-qi", pkg})
}

// InstallArgs builds SSH args for installing a package (confirm-gated).
func (m *Manager) InstallArgs(user, host, keyPath, pkg string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "dnf", []string{"install", "-y", pkg})
}

// UpdateArgs builds SSH args for updating a package (confirm-gated).
func (m *Manager) UpdateArgs(user, host, keyPath, pkg string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "dnf", []string{"update", "-y", pkg})
}

// ParseRpmList parses pipe-delimited rpm -qa output into Package structs.
func ParseRpmList(raw string) []Package {
	var pkgs []Package
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		pkgs = append(pkgs, Package{
			Name:    strings.TrimSpace(parts[0]),
			Version: strings.TrimSpace(parts[1]),
			Arch:    strings.TrimSpace(parts[2]),
		})
	}
	return pkgs
}

// IsMutating reports whether the operation modifies host state.
func IsMutating(op string) bool {
	return op == "install" || op == "update"
}

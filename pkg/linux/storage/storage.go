// Package storage provides Oracle Linux storage/LVM management over SSH.
package storage

import (
	"strings"

	"github.com/itunified-io/dbx/pkg/core/ssh"
)

// LogicalVolume represents an LVM logical volume.
type LogicalVolume struct {
	Name string
	VG   string
	Size string
}

// Manager provides storage operations.
type Manager struct {
	ssh *ssh.Executor
}

// New creates a storage manager with the given SSH executor.
func New(exec *ssh.Executor) *Manager {
	return &Manager{ssh: exec}
}

// PvListArgs builds SSH args for listing physical volumes.
func (m *Manager) PvListArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "pvs",
		[]string{"--noheadings", "--separator", "|", "-o", "pv_name,vg_name,pv_size,pv_free"})
}

// VgListArgs builds SSH args for listing volume groups.
func (m *Manager) VgListArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "vgs",
		[]string{"--noheadings", "--separator", "|", "-o", "vg_name,vg_size,vg_free,lv_count"})
}

// LvListArgs builds SSH args for listing logical volumes.
func (m *Manager) LvListArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "lvs",
		[]string{"--noheadings", "--separator", "|", "-o", "lv_name,vg_name,lv_size"})
}

// LvCreateArgs builds SSH args for creating a logical volume (confirm-gated).
func (m *Manager) LvCreateArgs(user, host, keyPath, name, size, vg string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "lvcreate", []string{"-L", size, "-n", name, vg})
}

// DiskUsageArgs builds SSH args for disk usage.
func (m *Manager) DiskUsageArgs(user, host, keyPath string) ([]string, error) {
	return m.ssh.BuildArgs(user, host, keyPath, "df", []string{"-hP"})
}

// ParseLvs parses pipe-delimited lvs output into LogicalVolume structs.
func ParseLvs(raw string) []LogicalVolume {
	var lvs []LogicalVolume
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(parts) != 3 {
			continue
		}
		lvs = append(lvs, LogicalVolume{
			Name: strings.TrimSpace(parts[0]),
			VG:   strings.TrimSpace(parts[1]),
			Size: strings.TrimSpace(parts[2]),
		})
	}
	return lvs
}

// IsMutating reports whether the operation modifies storage state.
func IsMutating(op string) bool {
	return op == "lv_create"
}

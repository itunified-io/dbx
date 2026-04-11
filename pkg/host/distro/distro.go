// Package distro provides a distro-agnostic abstraction layer for Linux host management.
package distro

// DistroID identifies a supported Linux distribution.
type DistroID int

const (
	Unknown     DistroID = iota
	Fedora               // Fedora 39+
	Ubuntu               // Ubuntu 22.04+ / Debian 12+
	RHEL                 // RHEL 8+ / CentOS Stream 8+
	SLES                 // SUSE Linux Enterprise Server 15+
	OracleLinux          // Oracle Linux 8+
)

func (d DistroID) String() string {
	switch d {
	case Fedora:
		return "fedora"
	case Ubuntu:
		return "ubuntu"
	case RHEL:
		return "rhel"
	case SLES:
		return "sles"
	case OracleLinux:
		return "oracle_linux"
	default:
		return "unknown"
	}
}

// Distro abstracts all distro-specific operations.
type Distro interface {
	ID() DistroID
	Name() string
	Version() string
	PackageManager() string
	FirewallTool() string
	InitSystem() string
	SELinuxAvailable() bool
	AppArmorAvailable() bool
	KspliceAvailable() bool

	ListPackagesCmd() []string
	ListUpdatesCmd(securityOnly bool) []string
	PackageInfoCmd(name string) []string
	ListFirewallRulesCmd() []string
	ServiceListCmd() []string
}

// NewAdapter returns the appropriate Distro implementation for the given ID.
func NewAdapter(id DistroID) Distro {
	switch id {
	case Fedora:
		return &fedoraAdapter{}
	case Ubuntu:
		return &ubuntuAdapter{}
	case RHEL:
		return &rhelAdapter{}
	case SLES:
		return &slesAdapter{}
	case OracleLinux:
		return &oracleLinuxAdapter{}
	default:
		return &rhelAdapter{}
	}
}

// baseAdapter provides common systemd-based defaults.
type baseAdapter struct {
	id      DistroID
	name    string
	version string
}

func (b *baseAdapter) ID() DistroID     { return b.id }
func (b *baseAdapter) Name() string     { return b.name }
func (b *baseAdapter) Version() string  { return b.version }
func (b *baseAdapter) InitSystem() string { return "systemd" }
func (b *baseAdapter) ServiceListCmd() []string {
	return []string{"systemctl", "list-units", "--type=service", "--all", "--no-pager"}
}

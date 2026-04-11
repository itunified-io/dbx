package distro

type fedoraAdapter struct{ baseAdapter }

func (f *fedoraAdapter) ID() DistroID         { return Fedora }
func (f *fedoraAdapter) Name() string         { return "Fedora" }
func (f *fedoraAdapter) Version() string      { return "" }
func (f *fedoraAdapter) PackageManager() string { return "dnf" }
func (f *fedoraAdapter) FirewallTool() string   { return "firewalld" }
func (f *fedoraAdapter) InitSystem() string     { return "systemd" }
func (f *fedoraAdapter) SELinuxAvailable() bool  { return true }
func (f *fedoraAdapter) AppArmorAvailable() bool { return false }
func (f *fedoraAdapter) KspliceAvailable() bool  { return false }

func (f *fedoraAdapter) ListPackagesCmd() []string {
	return []string{"dnf", "list", "installed", "--quiet"}
}
func (f *fedoraAdapter) ListUpdatesCmd(securityOnly bool) []string {
	if securityOnly {
		return []string{"dnf", "updateinfo", "list", "--security", "--quiet"}
	}
	return []string{"dnf", "check-update", "--quiet"}
}
func (f *fedoraAdapter) PackageInfoCmd(name string) []string {
	return []string{"dnf", "info", name}
}
func (f *fedoraAdapter) ListFirewallRulesCmd() []string {
	return []string{"firewall-cmd", "--list-all"}
}
func (f *fedoraAdapter) ServiceListCmd() []string {
	return []string{"systemctl", "list-units", "--type=service", "--all", "--no-pager"}
}

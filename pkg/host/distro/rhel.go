package distro

type rhelAdapter struct{ baseAdapter }

func (r *rhelAdapter) ID() DistroID         { return RHEL }
func (r *rhelAdapter) Name() string         { return "Red Hat Enterprise Linux" }
func (r *rhelAdapter) Version() string      { return "" }
func (r *rhelAdapter) PackageManager() string { return "dnf" }
func (r *rhelAdapter) FirewallTool() string   { return "firewalld" }
func (r *rhelAdapter) InitSystem() string     { return "systemd" }
func (r *rhelAdapter) SELinuxAvailable() bool  { return true }
func (r *rhelAdapter) AppArmorAvailable() bool { return false }
func (r *rhelAdapter) KspliceAvailable() bool  { return false }

func (r *rhelAdapter) ListPackagesCmd() []string {
	return []string{"dnf", "list", "installed", "--quiet"}
}
func (r *rhelAdapter) ListUpdatesCmd(securityOnly bool) []string {
	if securityOnly {
		return []string{"dnf", "updateinfo", "list", "--security", "--quiet"}
	}
	return []string{"dnf", "check-update", "--quiet"}
}
func (r *rhelAdapter) PackageInfoCmd(name string) []string {
	return []string{"dnf", "info", name}
}
func (r *rhelAdapter) ListFirewallRulesCmd() []string {
	return []string{"firewall-cmd", "--list-all"}
}
func (r *rhelAdapter) ServiceListCmd() []string {
	return []string{"systemctl", "list-units", "--type=service", "--all", "--no-pager"}
}

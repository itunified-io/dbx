package distro

type oracleLinuxAdapter struct{ baseAdapter }

func (o *oracleLinuxAdapter) ID() DistroID         { return OracleLinux }
func (o *oracleLinuxAdapter) Name() string         { return "Oracle Linux" }
func (o *oracleLinuxAdapter) Version() string      { return "" }
func (o *oracleLinuxAdapter) PackageManager() string { return "dnf" }
func (o *oracleLinuxAdapter) FirewallTool() string   { return "firewalld" }
func (o *oracleLinuxAdapter) InitSystem() string     { return "systemd" }
func (o *oracleLinuxAdapter) SELinuxAvailable() bool  { return true }
func (o *oracleLinuxAdapter) AppArmorAvailable() bool { return false }
func (o *oracleLinuxAdapter) KspliceAvailable() bool  { return true }

func (o *oracleLinuxAdapter) ListPackagesCmd() []string {
	return []string{"dnf", "list", "installed", "--quiet"}
}
func (o *oracleLinuxAdapter) ListUpdatesCmd(securityOnly bool) []string {
	if securityOnly {
		return []string{"dnf", "updateinfo", "list", "--security", "--quiet"}
	}
	return []string{"dnf", "check-update", "--quiet"}
}
func (o *oracleLinuxAdapter) PackageInfoCmd(name string) []string {
	return []string{"dnf", "info", name}
}
func (o *oracleLinuxAdapter) ListFirewallRulesCmd() []string {
	return []string{"firewall-cmd", "--list-all"}
}
func (o *oracleLinuxAdapter) ServiceListCmd() []string {
	return []string{"systemctl", "list-units", "--type=service", "--all", "--no-pager"}
}

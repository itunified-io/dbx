package distro

type slesAdapter struct{ baseAdapter }

func (s *slesAdapter) ID() DistroID         { return SLES }
func (s *slesAdapter) Name() string         { return "SUSE Linux Enterprise Server" }
func (s *slesAdapter) Version() string      { return "" }
func (s *slesAdapter) PackageManager() string { return "zypper" }
func (s *slesAdapter) FirewallTool() string   { return "firewalld" }
func (s *slesAdapter) InitSystem() string     { return "systemd" }
func (s *slesAdapter) SELinuxAvailable() bool  { return false }
func (s *slesAdapter) AppArmorAvailable() bool { return true }
func (s *slesAdapter) KspliceAvailable() bool  { return false }

func (s *slesAdapter) ListPackagesCmd() []string {
	return []string{"zypper", "se", "--installed-only", "--type=package"}
}
func (s *slesAdapter) ListUpdatesCmd(securityOnly bool) []string {
	if securityOnly {
		return []string{"zypper", "list-patches", "--category=security"}
	}
	return []string{"zypper", "list-updates"}
}
func (s *slesAdapter) PackageInfoCmd(name string) []string {
	return []string{"zypper", "info", name}
}
func (s *slesAdapter) ListFirewallRulesCmd() []string {
	return []string{"firewall-cmd", "--list-all"}
}
func (s *slesAdapter) ServiceListCmd() []string {
	return []string{"systemctl", "list-units", "--type=service", "--all", "--no-pager"}
}

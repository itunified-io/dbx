package distro

type ubuntuAdapter struct{ baseAdapter }

func (u *ubuntuAdapter) ID() DistroID         { return Ubuntu }
func (u *ubuntuAdapter) Name() string         { return "Ubuntu" }
func (u *ubuntuAdapter) Version() string      { return "" }
func (u *ubuntuAdapter) PackageManager() string { return "apt" }
func (u *ubuntuAdapter) FirewallTool() string   { return "ufw" }
func (u *ubuntuAdapter) InitSystem() string     { return "systemd" }
func (u *ubuntuAdapter) SELinuxAvailable() bool  { return false }
func (u *ubuntuAdapter) AppArmorAvailable() bool { return true }
func (u *ubuntuAdapter) KspliceAvailable() bool  { return false }

func (u *ubuntuAdapter) ListPackagesCmd() []string {
	return []string{"dpkg", "--list"}
}
func (u *ubuntuAdapter) ListUpdatesCmd(securityOnly bool) []string {
	if securityOnly {
		return []string{"apt", "list", "--upgradable", "-a"}
	}
	return []string{"apt", "list", "--upgradable"}
}
func (u *ubuntuAdapter) PackageInfoCmd(name string) []string {
	return []string{"apt-cache", "show", name}
}
func (u *ubuntuAdapter) ListFirewallRulesCmd() []string {
	return []string{"ufw", "status", "verbose"}
}
func (u *ubuntuAdapter) ServiceListCmd() []string {
	return []string{"systemctl", "list-units", "--type=service", "--all", "--no-pager"}
}

package distro

import (
	"fmt"
	"strings"
)

// ParseOSRelease parses /etc/os-release content and returns the appropriate Distro adapter.
func ParseOSRelease(content string) (Distro, error) {
	fields := parseKeyValue(content)

	id := strings.ToLower(fields["ID"])
	version := fields["VERSION_ID"]
	prettyName := fields["PRETTY_NAME"]

	var distroID DistroID
	switch id {
	case "ol":
		distroID = OracleLinux
	case "ubuntu", "debian":
		distroID = Ubuntu
	case "rhel", "centos", "rocky", "almalinux":
		distroID = RHEL
	case "fedora":
		distroID = Fedora
	case "sles", "sled", "opensuse-leap":
		distroID = SLES
	default:
		// Check ID_LIKE for fallback
		idLike := strings.ToLower(fields["ID_LIKE"])
		switch {
		case strings.Contains(idLike, "fedora") || strings.Contains(idLike, "rhel"):
			distroID = RHEL
		case strings.Contains(idLike, "debian"):
			distroID = Ubuntu
		case strings.Contains(idLike, "suse"):
			distroID = SLES
		default:
			return nil, fmt.Errorf("unsupported distribution: %s", id)
		}
	}

	adapter := NewAdapter(distroID)

	// Inject detected version and name into the adapter
	if da, ok := adapter.(interface{ SetDetected(string, string) }); ok {
		da.SetDetected(prettyName, version)
	}

	return &detectedAdapter{
		inner:   adapter,
		name:    prettyName,
		version: version,
	}, nil
}

// detectedAdapter wraps a base adapter with detected name/version.
type detectedAdapter struct {
	inner   Distro
	name    string
	version string
}

func (d *detectedAdapter) ID() DistroID             { return d.inner.ID() }
func (d *detectedAdapter) Name() string              { return d.name }
func (d *detectedAdapter) Version() string           { return d.version }
func (d *detectedAdapter) PackageManager() string    { return d.inner.PackageManager() }
func (d *detectedAdapter) FirewallTool() string      { return d.inner.FirewallTool() }
func (d *detectedAdapter) InitSystem() string        { return d.inner.InitSystem() }
func (d *detectedAdapter) SELinuxAvailable() bool     { return d.inner.SELinuxAvailable() }
func (d *detectedAdapter) AppArmorAvailable() bool    { return d.inner.AppArmorAvailable() }
func (d *detectedAdapter) KspliceAvailable() bool     { return d.inner.KspliceAvailable() }
func (d *detectedAdapter) ListPackagesCmd() []string  { return d.inner.ListPackagesCmd() }
func (d *detectedAdapter) ListUpdatesCmd(sec bool) []string { return d.inner.ListUpdatesCmd(sec) }
func (d *detectedAdapter) PackageInfoCmd(n string) []string { return d.inner.PackageInfoCmd(n) }
func (d *detectedAdapter) ListFirewallRulesCmd() []string   { return d.inner.ListFirewallRulesCmd() }
func (d *detectedAdapter) ServiceListCmd() []string         { return d.inner.ServiceListCmd() }

// parseKeyValue parses KEY=VALUE or KEY="VALUE" lines.
func parseKeyValue(content string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], "\"")
		result[key] = val
	}
	return result
}

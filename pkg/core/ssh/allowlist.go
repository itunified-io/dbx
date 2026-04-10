// Package ssh provides SSH command execution with an allowlist-based security model.
package ssh

// Allowlist maps domain -> set of allowed command basenames.
type Allowlist map[string]map[string]bool

// Has reports whether the given (domain, command) pair is allowed.
func (al Allowlist) Has(domain, command string) bool {
	cmds, ok := al[domain]
	if !ok {
		return false
	}
	return cmds[command]
}

// RegisterDomain merges a new domain + command list into an allowlist.
func RegisterDomain(al Allowlist, domain string, commands []string) {
	al[domain] = setOf(commands...)
}

var defaultAL = Allowlist{
	"patch":       setOf("opatch", "datapatch"),
	"backup":      setOf("rman"),
	"clusterware": setOf("crsctl", "srvctl", "olsnodes"),
	"asm":         setOf("asmcmd", "asmca"),
	"dataguard":   setOf("dgmgrl"),
	"rac":         setOf("srvctl"),
	"provision":   setOf("dbca", "netca"),
	"migration":   setOf("java"),
	"linux": setOf(
		// package management
		"rpm", "dnf",
		// kernel
		"sysctl", "cat", "uname",
		// storage / LVM
		"pvs", "vgs", "lvs", "lvcreate", "lvextend", "df", "lsblk", "multipath",
		// network
		"ip", "nmcli", "chronyc", "ss",
		// security
		"sestatus", "semanage", "firewall-cmd", "authconfig", "systemctl",
		// common
		"free",
	),
}

// DefaultAllowlist returns a copy of the process-wide allowlist.
func DefaultAllowlist() Allowlist {
	cp := make(Allowlist, len(defaultAL))
	for domain, cmds := range defaultAL {
		inner := make(map[string]bool, len(cmds))
		for k, v := range cmds {
			inner[k] = v
		}
		cp[domain] = inner
	}
	return cp
}

// RegisterDefaultDomain adds a domain to the process-wide default allowlist.
func RegisterDefaultDomain(domain string, commands []string) {
	defaultAL[domain] = setOf(commands...)
}

func setOf(items ...string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}

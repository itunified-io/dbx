package license

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
)

// GenerateFingerprint creates a deterministic machine fingerprint.
// SHA256(hostname + first-sorted-MAC + GOOS + GOARCH).
// Used for audit logging, not for phone-home.
func GenerateFingerprint() string {
	hostname, _ := os.Hostname()
	mac := firstMAC()
	input := fmt.Sprintf("%s|%s|%s|%s", hostname, mac, runtime.GOOS, runtime.GOARCH)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func firstMAC() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	var macs []string
	for _, iface := range ifaces {
		if iface.HardwareAddr != nil && len(iface.HardwareAddr) > 0 {
			macs = append(macs, iface.HardwareAddr.String())
		}
	}
	if len(macs) == 0 {
		return "unknown"
	}
	sort.Strings(macs)
	return macs[0]
}

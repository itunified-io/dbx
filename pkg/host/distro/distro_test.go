package distro_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/distro"
	"github.com/stretchr/testify/assert"
)

func TestDistroID_String(t *testing.T) {
	assert.Equal(t, "fedora", distro.Fedora.String())
	assert.Equal(t, "ubuntu", distro.Ubuntu.String())
	assert.Equal(t, "rhel", distro.RHEL.String())
	assert.Equal(t, "sles", distro.SLES.String())
	assert.Equal(t, "oracle_linux", distro.OracleLinux.String())
	assert.Equal(t, "unknown", distro.Unknown.String())
}

func TestDistroID_PackageManager(t *testing.T) {
	tests := []struct {
		id       distro.DistroID
		expected string
	}{
		{distro.Fedora, "dnf"},
		{distro.Ubuntu, "apt"},
		{distro.RHEL, "dnf"},
		{distro.SLES, "zypper"},
		{distro.OracleLinux, "dnf"},
	}
	for _, tt := range tests {
		d := distro.NewAdapter(tt.id)
		assert.Equal(t, tt.expected, d.PackageManager(), "distro=%s", tt.id)
	}
}

func TestDistroID_FirewallTool(t *testing.T) {
	tests := []struct {
		id       distro.DistroID
		expected string
	}{
		{distro.Fedora, "firewalld"},
		{distro.Ubuntu, "ufw"},
		{distro.RHEL, "firewalld"},
		{distro.SLES, "firewalld"},
		{distro.OracleLinux, "firewalld"},
	}
	for _, tt := range tests {
		d := distro.NewAdapter(tt.id)
		assert.Equal(t, tt.expected, d.FirewallTool(), "distro=%s", tt.id)
	}
}

func TestSELinuxAvailability(t *testing.T) {
	assert.True(t, distro.NewAdapter(distro.Fedora).SELinuxAvailable())
	assert.True(t, distro.NewAdapter(distro.RHEL).SELinuxAvailable())
	assert.True(t, distro.NewAdapter(distro.OracleLinux).SELinuxAvailable())
	assert.False(t, distro.NewAdapter(distro.Ubuntu).SELinuxAvailable())
	assert.False(t, distro.NewAdapter(distro.SLES).SELinuxAvailable())
}

func TestAppArmorAvailability(t *testing.T) {
	assert.True(t, distro.NewAdapter(distro.Ubuntu).AppArmorAvailable())
	assert.True(t, distro.NewAdapter(distro.SLES).AppArmorAvailable())
	assert.False(t, distro.NewAdapter(distro.Fedora).AppArmorAvailable())
}

func TestKspliceOnlyOracleLinux(t *testing.T) {
	assert.True(t, distro.NewAdapter(distro.OracleLinux).KspliceAvailable())
	assert.False(t, distro.NewAdapter(distro.Fedora).KspliceAvailable())
	assert.False(t, distro.NewAdapter(distro.Ubuntu).KspliceAvailable())
	assert.False(t, distro.NewAdapter(distro.RHEL).KspliceAvailable())
}

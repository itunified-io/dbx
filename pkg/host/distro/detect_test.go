package distro_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/distro"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOSRelease_OracleLinux(t *testing.T) {
	content := `NAME="Oracle Linux Server"
VERSION="8.9"
ID="ol"
ID_LIKE="fedora"
VERSION_ID="8.9"
PRETTY_NAME="Oracle Linux Server 8.9"
`
	d, err := distro.ParseOSRelease(content)
	require.NoError(t, err)
	assert.Equal(t, distro.OracleLinux, d.ID())
	assert.Equal(t, "8.9", d.Version())
}

func TestParseOSRelease_Ubuntu(t *testing.T) {
	content := `NAME="Ubuntu"
VERSION="22.04.4 LTS (Jammy Jellyfish)"
ID=ubuntu
ID_LIKE=debian
VERSION_ID="22.04"
PRETTY_NAME="Ubuntu 22.04.4 LTS"
`
	d, err := distro.ParseOSRelease(content)
	require.NoError(t, err)
	assert.Equal(t, distro.Ubuntu, d.ID())
	assert.Equal(t, "22.04", d.Version())
	assert.Equal(t, "apt", d.PackageManager())
}

func TestParseOSRelease_RHEL(t *testing.T) {
	content := `NAME="Red Hat Enterprise Linux"
VERSION="9.3 (Plow)"
ID="rhel"
VERSION_ID="9.3"
PRETTY_NAME="Red Hat Enterprise Linux 9.3 (Plow)"
`
	d, err := distro.ParseOSRelease(content)
	require.NoError(t, err)
	assert.Equal(t, distro.RHEL, d.ID())
	assert.Equal(t, "9.3", d.Version())
	assert.Equal(t, "dnf", d.PackageManager())
}

func TestParseOSRelease_Fedora(t *testing.T) {
	content := `NAME="Fedora Linux"
VERSION="41 (Workstation Edition)"
ID=fedora
VERSION_ID=41
PRETTY_NAME="Fedora Linux 41 (Workstation Edition)"
`
	d, err := distro.ParseOSRelease(content)
	require.NoError(t, err)
	assert.Equal(t, distro.Fedora, d.ID())
	assert.Equal(t, "dnf", d.PackageManager())
}

func TestParseOSRelease_SLES(t *testing.T) {
	content := `NAME="SLES"
VERSION="15-SP5"
ID="sles"
VERSION_ID="15.5"
PRETTY_NAME="SUSE Linux Enterprise Server 15 SP5"
`
	d, err := distro.ParseOSRelease(content)
	require.NoError(t, err)
	assert.Equal(t, distro.SLES, d.ID())
	assert.Equal(t, "zypper", d.PackageManager())
}

func TestParseOSRelease_FallbackIDLike(t *testing.T) {
	content := `NAME="Rocky Linux"
ID="rocky"
ID_LIKE="rhel centos fedora"
VERSION_ID="9.3"
PRETTY_NAME="Rocky Linux 9.3 (Blue Onyx)"
`
	d, err := distro.ParseOSRelease(content)
	require.NoError(t, err)
	assert.Equal(t, distro.RHEL, d.ID())
}

func TestParseOSRelease_Unsupported(t *testing.T) {
	content := `NAME="Arch Linux"
ID="arch"
PRETTY_NAME="Arch Linux"
`
	_, err := distro.ParseOSRelease(content)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

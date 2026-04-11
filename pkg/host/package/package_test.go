package hostpkg_test

import (
	"testing"

	hostpkg "github.com/itunified-io/dbx/pkg/host/package"
	"github.com/itunified-io/dbx/pkg/host/distro"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRpmList(t *testing.T) {
	output := `kernel|5.15.0-200.el8|x86_64
glibc|2.28-236.el8_9|x86_64
openssl|1.1.1k-12.el8_9|x86_64
`
	pkgs, err := hostpkg.ParsePackageList(output, distro.RHEL)
	require.NoError(t, err)
	assert.Len(t, pkgs, 3)
	assert.Equal(t, "kernel", pkgs[0].Name)
	assert.Equal(t, "5.15.0-200.el8", pkgs[0].Version)
	assert.Equal(t, "x86_64", pkgs[0].Arch)
}

func TestParseDpkgList(t *testing.T) {
	output := `ii  openssl      3.0.2-0ubuntu1.15  amd64  Secure Sockets Layer toolkit
ii  libc6:amd64  2.35-0ubuntu3.7    amd64  GNU C Library: Shared libraries
rc  old-pkg      1.0-1              amd64  Removed package
`
	pkgs, err := hostpkg.ParsePackageList(output, distro.Ubuntu)
	require.NoError(t, err)
	assert.Len(t, pkgs, 2)
	assert.Equal(t, "openssl", pkgs[0].Name)
	assert.Equal(t, "libc6", pkgs[1].Name)
}

func TestParseDnfUpdateInfo(t *testing.T) {
	output := `FEDORA-2026-abc123 Important/Sec. kernel-5.15.0-201.el8.x86_64
FEDORA-2026-def456 Moderate/Sec.  openssl-1.1.1k-13.el8_9.x86_64
FEDORA-2026-ghi789 bugfix         bash-4.4.20-5.el8.x86_64
`
	updates, err := hostpkg.ParseSecurityUpdates(output, distro.RHEL)
	require.NoError(t, err)
	assert.Len(t, updates, 3)
	assert.Equal(t, "Important", updates[0].Severity)
	assert.True(t, updates[0].Security)
	assert.False(t, updates[2].Security)
}

func TestParseAptUpdates(t *testing.T) {
	output := `openssl/jammy-security 3.0.2-0ubuntu1.16 amd64 [upgradable from: 3.0.2-0ubuntu1.15]
libc6/jammy-security 2.35-0ubuntu3.8 amd64 [upgradable from: 2.35-0ubuntu3.7]
`
	updates, err := hostpkg.ParseSecurityUpdates(output, distro.Ubuntu)
	require.NoError(t, err)
	assert.Len(t, updates, 2)
	assert.Equal(t, "openssl", updates[0].Name)
	assert.Equal(t, "3.0.2-0ubuntu1.16", updates[0].NewVersion)
	assert.Equal(t, "3.0.2-0ubuntu1.15", updates[0].OldVersion)
}

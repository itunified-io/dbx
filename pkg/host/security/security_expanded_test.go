package security_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSELinuxStatus(t *testing.T) {
	output := `SELinux status:                 enabled
SELinuxfs mount:                /sys/fs/selinux
SELinux root directory:         /etc/selinux
Loaded policy name:             targeted
Current mode:                   enforcing
Mode from config file:          enforcing
Policy MLS status:              enabled
`
	se, err := security.ParseSELinuxStatus(output)
	require.NoError(t, err)
	assert.True(t, se.Enabled)
	assert.Equal(t, "enforcing", se.Mode)
	assert.Equal(t, "targeted", se.Policy)
}

func TestParseAppArmorStatus(t *testing.T) {
	output := `apparmor module is loaded.
34 profiles are loaded.
34 profiles are in enforce mode.
0 profiles are in complain mode.
2 processes have profiles defined.
2 processes are in enforce mode.
0 processes are unconfined.
`
	aa, err := security.ParseAppArmorStatus(output)
	require.NoError(t, err)
	assert.True(t, aa.Loaded)
	assert.Equal(t, 34, aa.ProfilesEnforce)
	assert.Equal(t, 0, aa.ProfilesComplain)
	assert.Equal(t, 0, aa.Unconfined)
}

func TestSSHConfigCheck(t *testing.T) {
	content := `PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
Protocol 2
MaxAuthTries 3
X11Forwarding no
AllowUsers oracle dbmon
`
	checks := security.CheckSSHDConfig(content)
	assert.True(t, checks.RootLoginDisabled)
	assert.True(t, checks.PasswordAuthDisabled)
	assert.True(t, checks.PubkeyEnabled)
	assert.Equal(t, 3, checks.MaxAuthTries)
	assert.False(t, checks.X11Forwarding)
}

func TestSSHConfigInsecure(t *testing.T) {
	content := `PermitRootLogin yes
PasswordAuthentication yes
X11Forwarding yes
`
	checks := security.CheckSSHDConfig(content)
	assert.False(t, checks.RootLoginDisabled)
	assert.False(t, checks.PasswordAuthDisabled)
	assert.True(t, checks.X11Forwarding)
}

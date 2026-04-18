package security_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/security"
	"github.com/stretchr/testify/assert"
)

func TestParseSEStatus_Enforcing(t *testing.T) {
	content := `SELinux status:                 enabled
SELinuxfs mount:                /sys/fs/selinux
SELinux root directory:         /etc/selinux
Loaded policy name:             targeted
Current mode:                   enforcing
Mode from config file:          enforcing
Policy MLS status:              enabled
`
	s := security.ParseSEStatus(content)
	assert.True(t, s.Enabled)
	assert.Equal(t, "enforcing", s.Mode)
	assert.Equal(t, "targeted", s.Policy)
}

func TestParseSEStatus_Disabled(t *testing.T) {
	content := `SELinux status:                 disabled
`
	s := security.ParseSEStatus(content)
	assert.False(t, s.Enabled)
}

package service_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/service"
	"github.com/stretchr/testify/assert"
)

func TestParseSystemctlList(t *testing.T) {
	content := `  UNIT                       LOAD      ACTIVE   SUB     DESCRIPTION
  crond.service              loaded    active   running Command Scheduler
  dbus.service               loaded    active   running D-Bus System Message Bus
  firewalld.service          loaded    active   running firewalld - dynamic firewall daemon
  sshd.service               loaded    active   running OpenSSH server daemon
  tuned.service              loaded    active   running Dynamic System Tuning Daemon

LOAD   = Reflects whether the unit definition was properly loaded.
ACTIVE = The high-level unit activation state, i.e. generalization of SUB.
SUB    = The low-level unit activation state, values depend on unit type.

5 loaded units listed.
`
	units := service.ParseSystemctlList(content)
	assert.Len(t, units, 5)
	assert.Equal(t, "crond.service", units[0].Name)
	assert.Equal(t, "active", units[0].ActiveState)
	assert.Equal(t, "running", units[0].SubState)
}

func TestFailedUnits(t *testing.T) {
	units := []service.Unit{
		{Name: "sshd.service", ActiveState: "active", SubState: "running"},
		{Name: "bad.service", ActiveState: "failed", SubState: "failed"},
		{Name: "other.service", ActiveState: "inactive", SubState: "dead"},
	}
	failed := service.FailedUnits(units)
	assert.Len(t, failed, 1)
	assert.Equal(t, "bad.service", failed[0].Name)
}

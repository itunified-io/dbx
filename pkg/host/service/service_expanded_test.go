package service_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterByState(t *testing.T) {
	units := []service.Unit{
		{Name: "a.service", ActiveState: "active", SubState: "running"},
		{Name: "b.service", ActiveState: "failed", SubState: "failed"},
		{Name: "c.service", ActiveState: "active", SubState: "running"},
	}
	failed := service.FilterByState(units, "failed")
	assert.Len(t, failed, 1)
	assert.Equal(t, "b.service", failed[0].Name)
}

func TestParseSystemctlShow(t *testing.T) {
	output := `Type=notify
ActiveState=active
SubState=running
MainPID=12345
MemoryCurrent=83886080
ExecMainStartTimestamp=Wed 2026-04-10 09:00:00 UTC
Restart=on-failure
`
	detail, err := service.ParseSystemctlShow(output)
	require.NoError(t, err)
	assert.Equal(t, "notify", detail.Type)
	assert.Equal(t, 12345, detail.MainPID)
	assert.Equal(t, uint64(83886080), detail.MemoryBytes)
	assert.Equal(t, "on-failure", detail.RestartPolicy)
}

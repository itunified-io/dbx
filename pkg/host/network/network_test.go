package network_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/network"
	"github.com/stretchr/testify/assert"
)

func TestParseIPAddr(t *testing.T) {
	content := `lo               UNKNOWN        127.0.0.1/8 ::1/128
eth0             UP             10.0.0.10/24 fe80::1/64
docker0          DOWN
`
	ifaces := network.ParseIPAddr(content)
	assert.Len(t, ifaces, 3)
	assert.Equal(t, "lo", ifaces[0].Name)
	assert.Equal(t, "UNKNOWN", ifaces[0].State)
	assert.Contains(t, ifaces[0].Addrs, "127.0.0.1/8")
	assert.Equal(t, "eth0", ifaces[1].Name)
	assert.Equal(t, "UP", ifaces[1].State)
	assert.Equal(t, "docker0", ifaces[2].Name)
	assert.Equal(t, "DOWN", ifaces[2].State)
}

func TestParseSSListening(t *testing.T) {
	content := `State  Recv-Q  Send-Q  Local Address:Port  Peer Address:Port  Process
LISTEN 0       128     0.0.0.0:22           0.0.0.0:*          users:(("sshd",pid=1234,fd=3))
LISTEN 0       128     0.0.0.0:8080         0.0.0.0:*          users:(("java",pid=5678,fd=4))
`
	ports := network.ParseSSListening(content)
	assert.Len(t, ports, 2)
	assert.Equal(t, "22", ports[0].Port)
	assert.Equal(t, "8080", ports[1].Port)
}

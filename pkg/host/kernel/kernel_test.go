package kernel_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSysctl(t *testing.T) {
	output := `vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 3
kernel.shmmax = 68719476736
kernel.shmall = 4294967296
net.core.somaxconn = 65535
`
	params, err := kernel.ParseSysctl(output)
	require.NoError(t, err)
	assert.Equal(t, "10", params["vm.swappiness"])
	assert.Equal(t, "68719476736", params["kernel.shmmax"])
	assert.Equal(t, "65535", params["net.core.somaxconn"])
}

func TestParseHugepages(t *testing.T) {
	content := `HugePages_Total:    1024
HugePages_Free:      512
HugePages_Rsvd:      256
HugePages_Surp:        0
Hugepagesize:       2048 kB
`
	hp, err := kernel.ParseHugepages(content)
	require.NoError(t, err)
	assert.Equal(t, uint64(1024), hp.Total)
	assert.Equal(t, uint64(512), hp.Free)
	assert.Equal(t, uint64(256), hp.Reserved)
	assert.Equal(t, uint64(2048*1024), hp.SizeBytes)
	assert.InDelta(t, 50.0, hp.UsedPct, 1.0)
}

func TestParseUnameR(t *testing.T) {
	output := "5.15.0-200.el8.x86_64"
	info := kernel.ParseUnameR(output)
	assert.Equal(t, "5.15.0-200.el8.x86_64", info.Release)
	assert.Equal(t, "5.15.0", info.Version)
}

func TestParseLsmod(t *testing.T) {
	output := `Module                  Size  Used by
nf_conntrack          172032  3 nf_nat,nft_ct,xt_conntrack
ip_tables              28672  0
dm_mod                159744  2 dm_log,dm_mirror
`
	modules, err := kernel.ParseLsmod(output)
	require.NoError(t, err)
	assert.Len(t, modules, 3)
	assert.Equal(t, "nf_conntrack", modules[0].Name)
	assert.Equal(t, uint64(172032), modules[0].Size)
	assert.Equal(t, 3, modules[0].UsedBy)
}

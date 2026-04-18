package filesystem_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFindmnt(t *testing.T) {
	output := `TARGET        SOURCE    FSTYPE OPTIONS
/             /dev/sda1 ext4   rw,relatime
/boot         /dev/sda2 ext4   rw,relatime
/data         /dev/sdb1 xfs    rw,noatime,inode64
/mnt/nfs      nas:/vol1 nfs4   rw,hard,intr
`
	mounts, err := filesystem.ParseFindmnt(output)
	require.NoError(t, err)
	assert.Len(t, mounts, 4)

	data := mounts[2]
	assert.Equal(t, "/data", data.Target)
	assert.Equal(t, "/dev/sdb1", data.Source)
	assert.Equal(t, "xfs", data.FSType)
	assert.Contains(t, data.Options, "noatime")
}

func TestParseLVMLayout(t *testing.T) {
	pvsOutput := `  PV         VG       Fmt  Attr PSize   PFree
  /dev/sdb   data_vg  lvm2 a--  100.00g 20.00g
  /dev/sdc   data_vg  lvm2 a--  100.00g 50.00g
`
	lvsOutput := `  LV      VG       Attr       LSize  Pool Origin Data%  Meta%
  data_lv data_vg  -wi-ao---- 130.00g
  log_lv  data_vg  -wi-ao----  20.00g
`
	layout, err := filesystem.ParseLVMLayout(pvsOutput, lvsOutput)
	require.NoError(t, err)
	assert.Len(t, layout.PVs, 2)
	assert.Len(t, layout.LVs, 2)
	assert.Equal(t, "data_vg", layout.PVs[0].VG)
	assert.Equal(t, "data_lv", layout.LVs[0].Name)
}

func TestParseDfInodes(t *testing.T) {
	output := `Filesystem     Inodes  IUsed  IFree IUse% Mounted on
/dev/sda1      3276800 235000 3041800    8% /
/dev/sdb1      6553600 120000 6433600    2% /data
`
	inodes, err := filesystem.ParseDfInodes(output)
	require.NoError(t, err)
	assert.Len(t, inodes, 2)
	assert.Equal(t, uint64(3276800), inodes["/"].Total)
	assert.InDelta(t, 8.0, inodes["/"].UsedPct, 1.0)
}

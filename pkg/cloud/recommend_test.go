package cloud_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecommend_OracleOLTPSmall_AWS(t *testing.T) {
	rec, err := cloud.Recommend("oracle_oltp_small", cloud.AWS)
	require.NoError(t, err)
	assert.Equal(t, "r6i.xlarge", rec.InstanceType)
	assert.Equal(t, 4, rec.VCPUs)
	assert.Equal(t, 32, rec.MemoryGB)
	assert.Equal(t, 200, rec.StorageGB)
	assert.Equal(t, "io2", rec.StorageType)
	assert.Equal(t, 5000, rec.IOPS)
}

func TestRecommend_OracleOLTPMedium_Azure(t *testing.T) {
	rec, err := cloud.Recommend("oracle_oltp_medium", cloud.Azure)
	require.NoError(t, err)
	assert.Equal(t, "Standard_E8s_v5", rec.InstanceType)
	assert.Equal(t, 8, rec.VCPUs)
	assert.Equal(t, 64, rec.MemoryGB)
}

func TestRecommend_OracleOLTPLarge_OCI(t *testing.T) {
	rec, err := cloud.Recommend("oracle_oltp_large", cloud.OCI)
	require.NoError(t, err)
	assert.Equal(t, "VM.Standard.E4.Flex", rec.InstanceType)
	assert.Equal(t, 16, rec.VCPUs)
	assert.Equal(t, 256, rec.MemoryGB)
	assert.Equal(t, 1024, rec.StorageGB)
	assert.Equal(t, 20000, rec.IOPS)
}

func TestRecommend_PGOLTPSmall_AWS(t *testing.T) {
	rec, err := cloud.Recommend("pg_oltp_small", cloud.AWS)
	require.NoError(t, err)
	assert.Equal(t, "r6i.xlarge", rec.InstanceType)
	assert.Equal(t, 200, rec.StorageGB)
	assert.Equal(t, "gp3", rec.StorageType)
	assert.Equal(t, 3000, rec.IOPS)
}

func TestRecommend_PGOLTPLarge_Azure(t *testing.T) {
	rec, err := cloud.Recommend("pg_oltp_large", cloud.Azure)
	require.NoError(t, err)
	assert.Equal(t, "Standard_E16s_v5", rec.InstanceType)
	assert.Equal(t, 16, rec.VCPUs)
}

func TestRecommend_UnknownProfile(t *testing.T) {
	_, err := cloud.Recommend("unknown_profile", cloud.AWS)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown workload profile")
}

func TestRecommend_ListProfiles(t *testing.T) {
	profiles := cloud.ListWorkloadProfiles()
	assert.GreaterOrEqual(t, len(profiles), 5)
	assert.Contains(t, profiles, "oracle_oltp_small")
	assert.Contains(t, profiles, "pg_oltp_large")
}

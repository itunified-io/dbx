package cloud_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/cloud"
	"github.com/stretchr/testify/assert"
)

func TestProviderID_String(t *testing.T) {
	assert.Equal(t, "aws", cloud.AWS.String())
	assert.Equal(t, "azure", cloud.Azure.String())
	assert.Equal(t, "oci", cloud.OCI.String())
	assert.Equal(t, "onprem", cloud.OnPrem.String())
	assert.Equal(t, "unknown", cloud.ProviderID(99).String())
}

func TestInstanceState_String(t *testing.T) {
	assert.Equal(t, "running", cloud.StateRunning.String())
	assert.Equal(t, "stopped", cloud.StateStopped.String())
	assert.Equal(t, "terminated", cloud.StateTerminated.String())
	assert.Equal(t, "pending", cloud.StatePending.String())
}

func TestInstanceSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    cloud.InstanceSpec
		wantErr bool
	}{
		{
			name: "valid spec",
			spec: cloud.InstanceSpec{
				Name:         "ora-prod-01",
				InstanceType: "r6i.2xlarge",
				ImageID:      "ami-oracle-linux-8-latest",
				SubnetID:     "subnet-0abc123",
				Tags:         map[string]string{"Environment": "production"},
			},
			wantErr: false,
		},
		{
			name:    "empty name",
			spec:    cloud.InstanceSpec{InstanceType: "r6i.2xlarge", ImageID: "ami-123"},
			wantErr: true,
		},
		{
			name:    "empty instance type",
			spec:    cloud.InstanceSpec{Name: "test", ImageID: "ami-123"},
			wantErr: true,
		},
		{
			name:    "empty image",
			spec:    cloud.InstanceSpec{Name: "test", InstanceType: "r6i.2xlarge"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVolumeSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    cloud.VolumeSpec
		wantErr bool
	}{
		{
			name:    "valid volume",
			spec:    cloud.VolumeSpec{Name: "data", SizeGB: 500, Type: "io2", IOPS: 10000},
			wantErr: false,
		},
		{
			name:    "zero size",
			spec:    cloud.VolumeSpec{Name: "data", SizeGB: 0, Type: "io2"},
			wantErr: true,
		},
		{
			name:    "missing type",
			spec:    cloud.VolumeSpec{Name: "data", SizeGB: 500},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCostEstimate_MonthlyTotal(t *testing.T) {
	est := cloud.CostEstimate{
		ComputeMonthly: 1459.20,
		StorageMonthly: 987.60,
		NetworkMonthly: 66.20,
		ManagedDBMonthly: 0,
		Currency:       "USD",
	}
	assert.InDelta(t, 2513.00, est.MonthlyTotal(), 0.01)
}

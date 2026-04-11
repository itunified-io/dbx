package cloud_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRegistry_RegisterAndGet(t *testing.T) {
	reg := cloud.NewProviderRegistry()
	mock := &mockProvider{id: cloud.AWS}

	reg.Register(cloud.AWS, mock)

	got, err := reg.Get(cloud.AWS)
	require.NoError(t, err)
	assert.Equal(t, mock, got)
}

func TestProviderRegistry_GetUnregistered(t *testing.T) {
	reg := cloud.NewProviderRegistry()

	_, err := reg.Get(cloud.Azure)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no provider registered")
}

func TestProviderRegistry_List(t *testing.T) {
	reg := cloud.NewProviderRegistry()
	reg.Register(cloud.AWS, &mockProvider{id: cloud.AWS})
	reg.Register(cloud.OCI, &mockProvider{id: cloud.OCI})

	ids := reg.List()
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, cloud.AWS)
	assert.Contains(t, ids, cloud.OCI)
}

// mockProvider implements cloud.CloudProvider for testing.
type mockProvider struct {
	id cloud.ProviderID
}

func (m *mockProvider) CreateInstance(_ context.Context, spec cloud.InstanceSpec) (*cloud.Instance, error) {
	return &cloud.Instance{ID: "i-mock-001", Name: spec.Name, State: cloud.StatePending, Provider: m.id}, nil
}
func (m *mockProvider) TerminateInstance(_ context.Context, _ string) error { return nil }
func (m *mockProvider) GetInstance(_ context.Context, id string) (*cloud.Instance, error) {
	return &cloud.Instance{ID: id, State: cloud.StateRunning, Provider: m.id}, nil
}
func (m *mockProvider) ListInstances(_ context.Context, _ map[string]string) ([]*cloud.Instance, error) {
	return nil, nil
}
func (m *mockProvider) StartInstance(_ context.Context, _ string) error { return nil }
func (m *mockProvider) StopInstance(_ context.Context, _ string) error  { return nil }
func (m *mockProvider) CreateVolume(_ context.Context, spec cloud.VolumeSpec) (*cloud.Volume, error) {
	return &cloud.Volume{ID: "vol-mock-001", Name: spec.Name, SizeGB: spec.SizeGB}, nil
}
func (m *mockProvider) AttachVolume(_ context.Context, _, _ string) error { return nil }
func (m *mockProvider) DetachVolume(_ context.Context, _ string) error    { return nil }
func (m *mockProvider) CreateSecurityGroup(_ context.Context, spec cloud.SGSpec) (*cloud.SecurityGroup, error) {
	return &cloud.SecurityGroup{ID: "sg-mock-001", Name: spec.Name}, nil
}
func (m *mockProvider) AuthorizeIngress(_ context.Context, _ string, _ cloud.IngressRule) error {
	return nil
}
func (m *mockProvider) CreateLoadBalancer(_ context.Context, spec cloud.LBSpec) (*cloud.LoadBalancer, error) {
	return &cloud.LoadBalancer{ID: "lb-mock-001", Name: spec.Name}, nil
}
func (m *mockProvider) RegisterTarget(_ context.Context, _, _ string, _ int) error {
	return nil
}
func (m *mockProvider) CreateManagedDB(_ context.Context, spec cloud.ManagedDBSpec) (*cloud.ManagedDB, error) {
	return &cloud.ManagedDB{ID: "db-mock-001", Name: spec.Name, Engine: spec.Engine}, nil
}
func (m *mockProvider) GetManagedDB(_ context.Context, dbID string) (*cloud.ManagedDB, error) {
	return &cloud.ManagedDB{ID: dbID}, nil
}
func (m *mockProvider) EstimateCost(_ context.Context, _ cloud.InstanceSpec) (*cloud.CostEstimate, error) {
	return &cloud.CostEstimate{ComputeMonthly: 100.0, Currency: "USD"}, nil
}

package ha_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/ha"
	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func newFakeK8s() *pginternal.K8sClient {
	scheme := runtime.NewScheme()
	gvr := kschema.GroupVersionResource{Group: "postgresql.cnpg.io", Version: "v1", Resource: "clusters"}

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(kschema.GroupVersionKind{
		Group: "postgresql.cnpg.io", Version: "v1", Kind: "Cluster",
	})
	cluster.SetName("pg-cluster")
	cluster.SetNamespace("cnpg-system")

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[kschema.GroupVersionResource]string{gvr: "ClusterList"},
		cluster)

	return pginternal.NewK8sClientFromDynamic(client, "cnpg-system")
}

// ADR-0047: Failover must not proceed on a bare confirm boolean — the caller
// must restate the target cluster name via confirm_cluster.
func TestFailoverBooleanOnlyBlocks(t *testing.T) {
	k8s := newFakeK8s()
	_, err := ha.Failover(context.Background(), nil, k8s, map[string]any{
		"cluster": "pg-cluster",
		"target":  "pg-cluster-2",
		"confirm": true,
		// no confirm_cluster — must block
	})
	assert.ErrorContains(t, err, "identifier confirmation required")
}

func TestFailoverWrongClusterBlocks(t *testing.T) {
	k8s := newFakeK8s()
	_, err := ha.Failover(context.Background(), nil, k8s, map[string]any{
		"cluster":         "pg-cluster",
		"target":          "pg-cluster-2",
		"confirm":         true,
		"confirm_cluster": "wrong-cluster",
	})
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestFailoverCorrectRestatementPassesGate(t *testing.T) {
	k8s := newFakeK8s()
	result, err := ha.Failover(context.Background(), nil, k8s, map[string]any{
		"cluster":         "pg-cluster",
		"target":          "pg-cluster-2",
		"confirm":         true,
		"confirm_cluster": "pg-cluster",
	})
	require.NoError(t, err)
	assert.Equal(t, "pg-cluster-2", result.NewPrimary)
}

func TestSwitchoverRequiresConfirm(t *testing.T) {
	k8s := newFakeK8s()
	_, err := ha.Switchover(context.Background(), nil, k8s, map[string]any{
		"cluster": "pg-cluster",
		"target":  "pg-cluster-2",
		"confirm": false,
	})
	assert.ErrorContains(t, err, "confirm gate")
}

func TestSwitchoverSuccess(t *testing.T) {
	k8s := newFakeK8s()
	result, err := ha.Switchover(context.Background(), nil, k8s, map[string]any{
		"cluster": "pg-cluster",
		"target":  "pg-cluster-2",
		"confirm": true,
	})
	require.NoError(t, err)
	assert.Equal(t, "pg-cluster-2", result.NewPrimary)
}

func TestSwitchoverPlan(t *testing.T) {
	plan, err := ha.SwitchoverPlan(context.Background(), nil, nil, map[string]any{
		"target": "pg-cluster-2",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, plan.Steps)
	assert.Contains(t, plan.Steps[0], "pg-cluster-2")
}

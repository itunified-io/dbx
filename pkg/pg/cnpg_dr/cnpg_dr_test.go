package cnpg_dr_test

import (
	"context"
	"testing"

	cnpg_dr "github.com/itunified-io/dbx/pkg/pg/cnpg_dr"
	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func newFakeK8s(clusterName string) *pginternal.K8sClient {
	scheme := runtime.NewScheme()
	gvr := kschema.GroupVersionResource{Group: "postgresql.cnpg.io", Version: "v1", Resource: "clusters"}

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(kschema.GroupVersionKind{
		Group: "postgresql.cnpg.io", Version: "v1", Kind: "Cluster",
	})
	cluster.SetName(clusterName)
	cluster.SetNamespace("cnpg-system")
	cluster.Object["status"] = map[string]any{"phase": "Cluster in healthy state"}

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[kschema.GroupVersionResource]string{gvr: "ClusterList"},
		cluster)

	return pginternal.NewK8sClientFromDynamic(client, "cnpg-system")
}

func TestDRStatus(t *testing.T) {
	k8sPrimary := newFakeK8s("primary-cluster")
	k8sDR := newFakeK8s("dr-cluster")

	status, err := cnpg_dr.DRStatus(context.Background(), k8sPrimary, k8sDR, map[string]any{
		"primary": "primary-cluster",
		"dr":      "dr-cluster",
	})
	require.NoError(t, err)
	assert.Equal(t, "primary-cluster", status.PrimaryCluster)
	assert.True(t, status.InSync)
}

func TestDRPromoteRequiresDoubleConfirm(t *testing.T) {
	k8sDR := newFakeK8s("dr-cluster")
	_, err := cnpg_dr.DRPromote(context.Background(), k8sDR, map[string]any{
		"cluster":             "dr-cluster",
		"confirm":             true,
		"confirm_destructive": false,
	})
	assert.ErrorContains(t, err, "double-confirm required")
}

func TestDRSwitchoverRequiresDoubleConfirm(t *testing.T) {
	_, err := cnpg_dr.DRSwitchover(context.Background(), nil, nil, map[string]any{
		"primary":             "primary-cluster",
		"dr":                  "dr-cluster",
		"confirm":             true,
		"confirm_destructive": false,
	})
	assert.ErrorContains(t, err, "double-confirm required")
}

func TestFencingRequiresConfirm(t *testing.T) {
	k8s := newFakeK8s("pg-cluster")
	err := cnpg_dr.FencingEnable(context.Background(), k8s, "pg-cluster", false)
	assert.ErrorContains(t, err, "confirm gate")
}

func TestFencingEnableSuccess(t *testing.T) {
	k8s := newFakeK8s("pg-cluster")
	err := cnpg_dr.FencingEnable(context.Background(), k8s, "pg-cluster", true)
	require.NoError(t, err)
}

// ADR-0047: PITR recovery must not proceed on a bare confirm boolean — the caller
// must restate the target cluster name and PITR timestamp.
func TestRecoveryExecuteBooleanOnlyBlocks(t *testing.T) {
	_, err := cnpg_dr.RecoveryExecute(context.Background(), nil, "dr-cluster", "2026-04-10 14:30:00", true, "", "")
	assert.ErrorContains(t, err, "identifier confirmation required")
}

func TestRecoveryExecuteWrongClusterBlocks(t *testing.T) {
	_, err := cnpg_dr.RecoveryExecute(context.Background(), nil, "dr-cluster", "2026-04-10 14:30:00", true, "wrong-cluster", "2026-04-10 14:30:00")
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestRecoveryExecuteWrongTimestampBlocks(t *testing.T) {
	_, err := cnpg_dr.RecoveryExecute(context.Background(), nil, "dr-cluster", "2026-04-10 14:30:00", true, "dr-cluster", "2026-01-01 00:00:00")
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestRecoveryExecuteCorrectRestatementPassesGate(t *testing.T) {
	// With both factors correct, the gate is satisfied and the function proceeds past
	// the confirmation to its (stub) execution step.
	_, err := cnpg_dr.RecoveryExecute(context.Background(), nil, "dr-cluster", "2026-04-10 14:30:00", true, "dr-cluster", "2026-04-10 14:30:00")
	assert.ErrorContains(t, err, "recovery execution requires cluster configuration")
}

// ADR-0047: replica cluster deletion must not proceed on a bare confirm boolean —
// the caller must restate the target cluster name via confirm_cluster.
func TestReplicaClusterDeleteBooleanOnlyBlocks(t *testing.T) {
	err := cnpg_dr.ReplicaClusterDelete(context.Background(), nil, "replica-1", true, "")
	assert.ErrorContains(t, err, "identifier confirmation required")
}

func TestReplicaClusterDeleteWrongClusterBlocks(t *testing.T) {
	err := cnpg_dr.ReplicaClusterDelete(context.Background(), nil, "replica-1", true, "replica-2")
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestReplicaClusterDeleteCorrectClusterPassesGate(t *testing.T) {
	err := cnpg_dr.ReplicaClusterDelete(context.Background(), nil, "replica-1", true, "replica-1")
	assert.ErrorContains(t, err, "replica cluster deletion requires explicit cluster name")
}

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

func TestRecoveryExecuteRequiresDoubleConfirm(t *testing.T) {
	_, err := cnpg_dr.RecoveryExecute(context.Background(), nil, "", "2026-04-10 14:30:00", true, false)
	assert.ErrorContains(t, err, "double-confirm required")
}

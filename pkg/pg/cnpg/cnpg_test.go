package cnpg_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/cnpg"
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
	backupGVR := kschema.GroupVersionResource{Group: "postgresql.cnpg.io", Version: "v1", Resource: "backups"}

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(kschema.GroupVersionKind{
		Group: "postgresql.cnpg.io", Version: "v1", Kind: "Cluster",
	})
	cluster.SetName("pg-cluster")
	cluster.SetNamespace("cnpg-system")
	cluster.Object["spec"] = map[string]any{"instances": int64(3)}
	cluster.Object["status"] = map[string]any{
		"phase":          "Cluster in healthy state",
		"instances":      int64(3),
		"readyInstances": int64(3),
		"currentPrimary": "pg-cluster-1",
	}

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[kschema.GroupVersionResource]string{
			gvr:       "ClusterList",
			backupGVR: "BackupList",
		},
		cluster)

	return pginternal.NewK8sClientFromDynamic(client, "cnpg-system")
}

func TestClusterStatus(t *testing.T) {
	k8s := newFakeK8s()
	status, err := cnpg.ClusterStatus(context.Background(), k8s, "pg-cluster")
	require.NoError(t, err)
	assert.Equal(t, "pg-cluster", status.Name)
	assert.Contains(t, status.Phase, "healthy")
	assert.Equal(t, int64(3), status.ReadyInstances)
}

func TestClusterList(t *testing.T) {
	k8s := newFakeK8s()
	clusters, err := cnpg.ClusterList(context.Background(), k8s)
	require.NoError(t, err)
	assert.Len(t, clusters, 1)
	assert.Equal(t, "pg-cluster", clusters[0].Name)
}

func TestClusterScaleRequiresConfirm(t *testing.T) {
	k8s := newFakeK8s()
	err := cnpg.ClusterScale(context.Background(), k8s, "pg-cluster", 5, false)
	assert.ErrorContains(t, err, "confirm gate")
}

func TestClusterScaleSuccess(t *testing.T) {
	k8s := newFakeK8s()
	err := cnpg.ClusterScale(context.Background(), k8s, "pg-cluster", 5, true)
	require.NoError(t, err)
}

func TestBackupTriggerRequiresConfirm(t *testing.T) {
	k8s := newFakeK8s()
	_, err := cnpg.BackupTrigger(context.Background(), k8s, "pg-cluster", false)
	assert.ErrorContains(t, err, "confirm gate")
}

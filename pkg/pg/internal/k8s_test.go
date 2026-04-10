package pginternal_test

import (
	"context"
	"testing"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestGetCluster(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := kschema.GroupVersionResource{Group: "postgresql.cnpg.io", Version: "v1", Resource: "clusters"}

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(kschema.GroupVersionKind{
		Group: "postgresql.cnpg.io", Version: "v1", Kind: "Cluster",
	})
	cluster.SetName("pg-prod-cluster")
	cluster.SetNamespace("cnpg-system")
	cluster.Object["spec"] = map[string]any{"instances": int64(3)}
	cluster.Object["status"] = map[string]any{
		"phase":          "Cluster in healthy state",
		"instances":      int64(3),
		"readyInstances": int64(3),
	}

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[kschema.GroupVersionResource]string{gvr: "ClusterList"},
		cluster)

	k8s := pginternal.NewK8sClientFromDynamic(client, "cnpg-system")
	result, err := k8s.GetCluster(context.Background(), "pg-prod-cluster")
	require.NoError(t, err)
	assert.Equal(t, "pg-prod-cluster", result.GetName())

	status, _, _ := unstructured.NestedString(result.Object, "status", "phase")
	assert.Contains(t, status, "healthy")
}

func TestListClusters(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := kschema.GroupVersionResource{Group: "postgresql.cnpg.io", Version: "v1", Resource: "clusters"}

	c1 := &unstructured.Unstructured{}
	c1.SetGroupVersionKind(kschema.GroupVersionKind{Group: "postgresql.cnpg.io", Version: "v1", Kind: "Cluster"})
	c1.SetName("cluster-a")
	c1.SetNamespace("cnpg-system")

	c2 := &unstructured.Unstructured{}
	c2.SetGroupVersionKind(kschema.GroupVersionKind{Group: "postgresql.cnpg.io", Version: "v1", Kind: "Cluster"})
	c2.SetName("cluster-b")
	c2.SetNamespace("cnpg-system")

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[kschema.GroupVersionResource]string{gvr: "ClusterList"},
		c1, c2)

	k8s := pginternal.NewK8sClientFromDynamic(client, "cnpg-system")
	clusters, err := k8s.ListClusters(context.Background())
	require.NoError(t, err)
	assert.Len(t, clusters.Items, 2)
}

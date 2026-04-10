package pginternal

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var cnpgGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Version:  "v1",
	Resource: "clusters",
}

var cnpgBackupGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Version:  "v1",
	Resource: "backups",
}

// K8sClient wraps the Kubernetes dynamic client for CNPG operations.
type K8sClient struct {
	client    dynamic.Interface
	namespace string
}

// NewK8sClientFromDynamic creates a K8sClient from an existing dynamic.Interface (for testing).
func NewK8sClientFromDynamic(client dynamic.Interface, namespace string) *K8sClient {
	return &K8sClient{client: client, namespace: namespace}
}

// GetCluster retrieves a CNPG Cluster by name.
func (k *K8sClient) GetCluster(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	result, err := k.client.Resource(cnpgGVR).Namespace(k.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get cluster %s: %w", name, err)
	}
	return result, nil
}

// ListClusters lists all CNPG Clusters in the namespace.
func (k *K8sClient) ListClusters(ctx context.Context) (*unstructured.UnstructuredList, error) {
	result, err := k.client.Resource(cnpgGVR).Namespace(k.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	return result, nil
}

// PatchCluster applies a merge patch to a CNPG Cluster.
func (k *K8sClient) PatchCluster(ctx context.Context, name string, patchData []byte) error {
	_, err := k.client.Resource(cnpgGVR).Namespace(k.namespace).Patch(
		ctx, name, types.MergePatchType, patchData, metav1.PatchOptions{})
	return err
}

// ListBackups lists CNPG Backup resources.
func (k *K8sClient) ListBackups(ctx context.Context, labelSelector string) (*unstructured.UnstructuredList, error) {
	opts := metav1.ListOptions{}
	if labelSelector != "" {
		opts.LabelSelector = labelSelector
	}
	result, err := k.client.Resource(cnpgBackupGVR).Namespace(k.namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	return result, nil
}

// CreateBackup creates a CNPG Backup resource.
func (k *K8sClient) CreateBackup(ctx context.Context, backup *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	result, err := k.client.Resource(cnpgBackupGVR).Namespace(k.namespace).Create(ctx, backup, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create backup: %w", err)
	}
	return result, nil
}

// Namespace returns the configured namespace.
func (k *K8sClient) Namespace() string {
	return k.namespace
}
